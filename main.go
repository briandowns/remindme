package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
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

const remindmeSock = "/tmp/remindme.sock"

var validDest = []string{"at", "in", "on"}

// notification
type notification struct {
	id   int
	spec string
	text string
	c    *cron.Cron
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

		n.spec = minute + " " + hour + " * * *"
	case "in":
		now := time.Now()

		dur, err := time.ParseDuration(parts[1])
		if err != nil {
			return nil, err
		}

		switch {
		case strings.Contains(parts[1], "h"):
			h := int(dur.Hours()) + now.Hour()
			n.spec = fmt.Sprintf("* %d * * *", h)
		case strings.Contains(parts[1], "m"):
			m := int(dur.Minutes()) + now.Minute()
			n.spec = fmt.Sprintf("%d * * * *", m)
		}
	case "on":
		on := strings.Split(parts[1], "/")
		month, day := on[0], on[1]

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

		if err := os.RemoveAll(remindmeSock); err != nil {
			logger.Fatal(err.Error())
		}

		logger.Info("initializing internal cron service")
		l, err := net.Listen("unix", remindmeSock)
		if err != nil {
			logger.Fatal(err.Error())
		}
		defer l.Close()

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

			entryID, err := c.AddJob(n.spec, n)
			if err != nil {
				logger.Fatal(err.Error())
			}
			n.id = int(entryID)

			logger.Info("new reminder scheduled", zap.Int("id", n.id))
		}
	}

	if len(os.Args) != 4 {
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
