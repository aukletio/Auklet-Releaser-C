package main

import (
	"fmt"
	"os"
	"os/exec"
)

// Run a command and report when it exits.
func run(cmd *exec.Cmd, done chan struct{}) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	check(err)
	err = cmd.Wait()
	fmt.Println(err)

	// There is a race condition between child exit and socket EOF; we don't
	// know which will happen first.  Nevertheless, if the child exits,
	// there is no longer any stdout or stderr to print out (those files
	// should close). run() should report the exit status of the command.
	done <- struct{}{}
}
