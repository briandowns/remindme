package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

var (
	name    string
	version string
	gitSHA  string
)

const remindmeSock = "/tmp/remindme.sock"

var validDest = []string{"at", "in", "on"}

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
Usage: %[2]s [-v] [-h]
Options:
    -h        help
    -v        show version and exit
Examples: 
    %[2]s at 09:16 "call the handyman"
	%[2]s in 5m "login to the meeting"
	%[2]s on 08/17 "buy a birthday card"
`

func main() {
	var vers bool

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
	flag.Parse()

	if vers {
		fmt.Fprintf(os.Stdout, "version: %s - git sha: %s\n", version, gitSHA)
		return
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
