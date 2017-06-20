package main

import (
	"fmt"
	"bufio"
	"os/exec"
)

func run(cmd *exec.Cmd, quit chan struct{}) {
	stdout, err := cmd.StdoutPipe()
	check(err)

	stderr, err := cmd.StderrPipe()
	check(err)

	err = cmd.Start()
	check(err)

	go func() {
		s := bufio.NewScanner(stdout)
		for s.Scan() {
			fmt.Println(s.Text())
		}
	}()

	go func() {
		s := bufio.NewScanner(stderr)
		for s.Scan() {
			fmt.Println(s.Text())
		}
	}()

	err = cmd.Wait()
	check(err)
	quit <- struct{}{}
}
