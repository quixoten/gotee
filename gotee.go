package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	var flagAppend bool
	var filePath string
	cmd := flag.NewFlagSet("gotee", flag.ExitOnError)
	openFlags := os.O_WRONLY|os.O_CREATE|os.O_TRUNC

	cmd.BoolVar(&flagAppend, "a", false, "Append to FILE. Do not overwrite.")
	cmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: gotee [options] FILE\n")
		cmd.PrintDefaults()
	}
	cmd.Parse(os.Args[1:])
	filePath = cmd.Arg(0)

	if filePath == "" {
		cmd.Usage()
		os.Exit(2)
	}

	if flagAppend {
		openFlags = os.O_WRONLY|os.O_CREATE|os.O_APPEND
	}

	file, err := os.OpenFile(filePath, openFlags, 0666)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	buffer := make([]byte, 4096)
	writer := io.MultiWriter(file, os.Stdout)
	tee := io.TeeReader(os.Stdin, writer)
	var info1, info2 os.FileInfo

	for {
		info1, err = os.Stat(filePath)
		info2, err = file.Stat()
		if !os.SameFile(info1, info2) {
			file.Close()
			file, err = os.OpenFile(filePath, openFlags, 0666)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			writer = io.MultiWriter(file, os.Stdout)
			tee = io.TeeReader(os.Stdin, writer)
			fmt.Fprintln(os.Stderr, "log file re-opened")
		}

		_, err := tee.Read(buffer)
		if err == nil {
			continue
		}
		if err == io.EOF {
			break
		}

		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	file.Close()
}
