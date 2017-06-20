package main

import (
	"fmt"
	"os/exec"
)

func run(cmd *exec.Cmd, quit chan int) {
	log, err := cmd.CombinedOutput()

	fmt.Print(err, string(log))
	quit <- 0
}
