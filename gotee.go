package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

var usage = `Gotee is a version of the tee program that re-opens its output file when it is
moved, re-created, or deleted by an external program, such as logrotate.

Usage:
  gotee [options] FILE

Options:
  -a           Append to FILE. Do not truncate on open/re-open.
  -i duration  Interval between checking for changes to FILE. Default is 5s.
  -v           Show version and exit.
`

type Cmd struct {
	in            io.Reader
	out           io.Writer
	err           io.Writer
	logger        *log.Logger
	path          string
	append        bool
	checkInterval time.Duration
	showVersion   bool
	flags         *flag.FlagSet
}

var version = "0.0.1"

func main() {
	cmd := New()

	if err := cmd.Parse(os.Args[1:]); err != nil {
		cmd.flags.Usage()
		os.Exit(1)
	}

	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func New() (cmd *Cmd) {
	cmd = &Cmd{
		in:            os.Stdin,
		out:           os.Stdout,
		err:           os.Stderr,
		path:          "",
		append:        false,
		checkInterval: 1 * time.Second,
		showVersion:   false,
		flags:         flag.NewFlagSet("gotee", flag.ContinueOnError),
	}

	cmd.logger = log.New(cmd.err, "gotee: ", log.LstdFlags)
	cmd.flags.BoolVar(&cmd.append, "a", false, "")
	cmd.flags.DurationVar(&cmd.checkInterval, "i", 5*time.Second, "")
	cmd.flags.BoolVar(&cmd.showVersion, "v", false, "")
	cmd.flags.Usage = func() {
		fmt.Fprintln(os.Stderr, usage)
	}

	return cmd
}

func (cmd *Cmd) Parse(args []string) error {
	if err := cmd.flags.Parse(args); err != nil {
		return err
	}

	cmd.path = cmd.flags.Arg(0)

	return nil
}

func (cmd *Cmd) Run() error {
	if cmd.showVersion {
		fmt.Fprintf(cmd.err, "gotee version %s\n", version)
		return nil
	}

	if cmd.path == "" {
		return fmt.Errorf("FILE not specified.\n\n%s\n", usage)
	}

	openFile, openFileInfo, err := cmd.openFile()
	if err != nil {
		return err
	}

	buffer := make([]byte, 4096)
	writer := io.MultiWriter(openFile, cmd.out)
	nextCheckForChanges := time.Now().Add(cmd.checkInterval)

	for {
		bytesRead, err := cmd.in.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		now := time.Now()
		if now.After(nextCheckForChanges) {
			nextCheckForChanges = now.Add(cmd.checkInterval)
			onDiskFileInfo, _ := os.Stat(cmd.path)
			if !os.SameFile(openFileInfo, onDiskFileInfo) {
				openFile.Close()
				openFile, openFileInfo, err = cmd.openFile()
				if err != nil {
					return err
				}
				writer = io.MultiWriter(openFile, cmd.out)
				cmd.logger.Printf("re-opened output file '%s'\n", cmd.path)
			}
		}

		_, err = writer.Write(buffer[:bytesRead])
		if err != nil {
			return err
		}
	}

	return nil
}

func (cmd *Cmd) openFile() (file *os.File, info os.FileInfo, err error) {
	openFlags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if cmd.append {
		openFlags = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	}

	file, err = os.OpenFile(cmd.path, openFlags, 0666)
	if err != nil {
		return
	}
	info, err = file.Stat()
	if err != nil {
		return
	}

	return
}
