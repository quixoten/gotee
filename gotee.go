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
moved or re-created by an external program, such as logrotate.

Usage:
  gotee [options] FILE

Options:
  -a           Append to FILE. Do not truncate on open/re-open.
  -i duration  Interval between checking for changes to FILE. Default is 5s.
`

type Tee struct {
	in            io.Reader
	out           io.Writer
	logger        *log.Logger
	path          string
	append        bool
	checkInterval time.Duration
}

func main() {
	tee := &Tee{
		in:            os.Stdin,
		out:           os.Stdout,
		logger:        log.New(os.Stderr, "gotee: ", log.LstdFlags),
		append:        false,
		checkInterval: 1 * time.Second,
	}

	cmd := flag.NewFlagSet("gotee", flag.ExitOnError)
	cmd.BoolVar(&tee.append, "a", false, "Append to FILE. Do not overwrite.")
	cmd.DurationVar(&tee.checkInterval, "i", 5*time.Second, "Interval between checks for changes to FILE.")
	cmd.Usage = func() {
		fmt.Fprintln(os.Stderr, usage)
	}
	cmd.Parse(os.Args[1:])

	tee.path = cmd.Arg(0)
	if tee.path == "" {
		cmd.Usage()
		os.Exit(2)
	}

	if err := tee.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func (t *Tee) openFile() (file *os.File, info os.FileInfo, err error) {
	openFlags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if t.append {
		openFlags = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	}

	file, err = os.OpenFile(t.path, openFlags, 0666)
	if err != nil {
		return
	}
	info, err = file.Stat()
	if err != nil {
		return
	}

	return
}

func (t *Tee) Run() error {
	openFile, openFileInfo, err := t.openFile()
	if err != nil {
		return err
	}

	buffer := make([]byte, 4096)
	writer := io.MultiWriter(openFile, t.out)
	nextCheckForChanges := time.Now().Add(t.checkInterval)

	for {
		bytesRead, err := t.in.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		now := time.Now()
		if now.After(nextCheckForChanges) {
			nextCheckForChanges = now.Add(t.checkInterval)
			onDiskFileInfo, _ := os.Stat(t.path)
			if !os.SameFile(openFileInfo, onDiskFileInfo) {
				openFile.Close()
				openFile, openFileInfo, err = t.openFile()
				if err != nil {
					return err
				}
				writer = io.MultiWriter(openFile, t.out)
				t.logger.Printf("re-opened output file '%s'\n", t.path)
			}
		}

		_, err = writer.Write(buffer[:bytesRead])
		if err != nil {
			return err
		}
	}

	return nil
}
