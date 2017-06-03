package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

func main() {
	var cmdAppend bool

	cmd := flag.NewFlagSet("gotee", flag.ExitOnError)
	cmd.BoolVar(&cmdAppend, "a", false, "Append to FILE. Do not overwrite.")
	cmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: gotee [options] FILE\n")
		cmd.PrintDefaults()
	}
	cmd.Parse(os.Args[1:])

	filePath := cmd.Arg(0)
	if filePath == "" {
		cmd.Usage()
		os.Exit(2)
	}

	openFlags := os.O_WRONLY|os.O_CREATE|os.O_TRUNC
	if cmdAppend {
		openFlags = os.O_WRONLY|os.O_CREATE|os.O_APPEND
	}

	file, err := os.OpenFile(filePath, openFlags, 0666)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	buffer := make([]byte, 4096)
	writer := io.MultiWriter(file, os.Stdout)
	nextVanishCheck := time.Now().Add(5 * time.Second)

	for {
		bytesRead, err := os.Stdin.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		now := time.Now()
		if now.After(nextVanishCheck) {
			nextVanishCheck = now.Add(5 * time.Second)
			if _, err := os.Stat(filePath); err != nil {
				file.Close()
				file, err = os.OpenFile(filePath, openFlags, 0666)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				writer = io.MultiWriter(file, os.Stdout)
				fmt.Fprintln(os.Stderr, "log file re-opened")
			}
		}

		_, err = writer.Write(buffer[:bytesRead])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
