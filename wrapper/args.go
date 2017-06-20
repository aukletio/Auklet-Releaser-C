package main

import (
	"fmt"
	"os"
	"os/exec"
)

func usage() {
	fmt.Printf("usage: %v command [args ...]\n", os.Args[0])
	os.Exit(0)
}

func command() *exec.Cmd {
	if len(os.Args) < 2 {
		usage()
	}

	return exec.Command(os.Args[1], os.Args[2:]...)
}
