package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
)

// Run a command and report when it exits.
func run(cmd *exec.Cmd, wg sync.WaitGroup) {
	defer wg.Done()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	check(err)
	err = cmd.Wait()
	fmt.Println(err)
}
