package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
)

// Run a command and report when it exits.
func run(cmd *exec.Cmd, done chan struct{}) {
	// Set up pipes so we can see the command's text output.
	stdout, err := cmd.StdoutPipe()
	check(err)

	stderr, err := cmd.StderrPipe()
	check(err)

	thru := func(f io.ReadCloser) {
		s := bufio.NewScanner(f)

		// TODO: Route stdout and stderr to channels so they can be
		// accumulated in a profile.

		for s.Scan() {
			fmt.Println(s.Text())
		}
	}

	err = cmd.Start()
	check(err)

	go thru(stderr)
	go thru(stdout)

	err = cmd.Wait()

	// There is a race condition between child exit and socket EOF; we don't
	// know which will happen first.  Nevertheless, if the child exits,
	// there is no longer any stdout or stderr to print out (those files
	// should close). run() should report the exit status of the command.
	done <- struct{}{}
}
