package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/martinlindhe/notify"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

var (
	name    string
	version string
	gitSHA  string
)

const usage = `version: %s
Usage: %[2]s [-v] [-h] [-s]
Options:
    -s        run the server
    -h        help
    -v        show version and exit
Examples:
    %[2]s at 09:16 "call the handyman"
    %[2]s in 5m "login to the meeting"
    %[2]s on 08/17 "buy a birthday card"
`

const (
	remindmeSock = "/tmp/remindme.sock"
	remindmePid = "/tmp/remindme.pid"

	maxArgs = 4
)

var validDest = []string{"at", "in", "on"}

const (
	timerNotification = iota
	cronNotification
)

// notification
type notification struct {
	notificationType int
	id               int
	spec             string
	text             string
	c                *cron.Cron
	dur              time.Duration
}

// Run performs the operation of
func (n notification) Run() {
	notify.Alert("remindme", "", n.text, "")
	n.c.Remove(cron.EntryID(n.id))
}

// parse reads the bytes from the connection and parses
// it out into a "notification".
func parse(conn net.Conn, c *cron.Cron) (*notification, error) {
	b := bytes.NewBuffer(nil)
	io.Copy(b, conn)

	parts := strings.Split(b.String(), "|")

	n := notification{
		text: parts[2],
		c:    c,
	}

	switch parts[0] {
	case "at":
		at := strings.Split(parts[1], ":")
		hour, minute := at[0], at[1]

		n.notificationType = cronNotification
		n.spec = minute + " " + hour + " * * *"
	case "in":
		dur, err := time.ParseDuration(parts[1])
		if err != nil {
			return nil, err
		}

		n.notificationType = timerNotification
		n.dur = dur
	case "on":
		on := strings.Split(parts[1], "/")
		month, day := on[0], on[1]

		n.notificationType = cronNotification
		n.spec = "* * " + day + " " + month + " * *"
	default:
		return nil, errors.New("invalid descriptor")
	}

	return &n, nil
}

// validate checks what is given by the user from
// the terminal and makes sure it's able to be
// successfully processed.
func validate(args []string) error {
	for _, d := range validDest {
		if args[1] == d {
			break
		}
	}

	if args[1] == "at" && !strings.Contains(args[2], ":") {
		return errors.New("error: invalid format for 'at'")
	}

	if args[1] == "in" {
		_, err := time.ParseDuration(args[2])
		if err != nil {
			return err
		}

		if !strings.Contains(args[2], "m") && !strings.Contains(args[2], "h") {
			return errors.New("error: unsupported duration. Only 'm' or 'h' currently supported")
		}
	}

	if args[1] == "on" && !strings.Contains(args[2], "/") {
		return errors.New("error: invalid format for 'on'")
	}

	return nil
}

// stopCurrentProcess check if a current server-process is running
// if it exists, the existing process is terminated
func stopCurrentProcess(pathToPidFile string) error {
	pidData, err := os.ReadFile(pathToPidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	pid, err := strconv.Atoi(string(pidData))
	if err != nil {
		return err
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	err = process.Signal(syscall.Signal(syscall.SIGINT))
	if err != nil {
		if errors.Is(err, os.ErrProcessDone) {
			return nil
		}
		return err
	}

	return nil
}

func main() {
	var vers bool
	var server bool

	flag.Usage = func() {
		w := os.Stderr
		for _, arg := range os.Args {
			if arg == "-h" {
				w = os.Stdout
				break
			}
		}
		fmt.Fprintf(w, usage, version, name)
	}

	flag.BoolVar(&vers, "v", false, "")
	flag.BoolVar(&server, "s", false, "")
	flag.Parse()

	if vers {
		fmt.Fprintf(os.Stdout, "version: %s - git sha: %s\n", version, gitSHA)
		return
	}

	if server {
		logger, err := zap.NewProduction()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer logger.Sync()

		if err := stopCurrentProcess(remindmePid); err != nil {
			logger.Fatal(err.Error())
		}

		if err := os.RemoveAll(remindmePid); err != nil {
			logger.Fatal(err.Error())
		}

		if err := os.RemoveAll(remindmeSock); err != nil {
			logger.Fatal(err.Error())
		}

		logger.Info("initializing internal cron service")
		l, err := net.Listen("unix", remindmeSock)
		if err != nil {
			logger.Fatal(err.Error())
		}
		defer l.Close()

		pid := syscall.Getpid()

		err = os.WriteFile(remindmePid, []byte(fmt.Sprintf("%d", pid)), 0644)
		if err != nil {
			logger.Fatal(err.Error())
		}

		logger.Info("starting internal cron service")
		c := cron.New()
		go func() {
			c.Start()
		}()

		logger.Info("accepting connections")
		for {
			conn, err := l.Accept()
			if err != nil {
				logger.Fatal(err.Error())
			}

			n, err := parse(conn, c)
			if err != nil {
				logger.Warn(err.Error())
				continue
			}

			switch n.notificationType {
			case cronNotification:
				entryID, err := c.AddJob(n.spec, n)
				if err != nil {
					logger.Fatal(err.Error())
				}
				n.id = int(entryID)
			case timerNotification:
				go func() {
					t := time.NewTimer(n.dur)
					<-t.C
					n.Run()
				}()

			}

			logger.Info("new reminder scheduled", zap.Int("id", n.id))
		}
	}

	if len(os.Args) != maxArgs {
		fmt.Fprintf(os.Stderr, usage, version, name)
		os.Exit(1)
	}

	if err := validate(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	c, err := net.Dial("unix", remindmeSock)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if _, err := c.Write([]byte(os.Args[1] + "|" + os.Args[2] + "|" + os.Args[3])); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
