package main

import (
	"fmt"
	"os/exec"
)

func run(cmd *exec.Cmd, quit chan struct{}) {
	log, err := cmd.CombinedOutput()

	fmt.Print(err, string(log))
	quit <- struct{}{}
}
