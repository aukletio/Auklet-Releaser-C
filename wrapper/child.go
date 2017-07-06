package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
)

func run(cmd *exec.Cmd, quit chan struct{}) {
	stdout, err := cmd.StdoutPipe()
	check(err)

	stderr, err := cmd.StderrPipe()
	check(err)

	thru := func(f io.ReadCloser) {
		s := bufio.NewScanner(f)

		for s.Scan() {
			fmt.Println(s.Text())
		}
	}

	defer func() {
		if x := recover(); x != nil {
			fmt.Println(x)
		}
	}()

	err = cmd.Start()
	check(err)

	go thru(stderr)
	go thru(stdout)

	err = cmd.Wait()
	check(err)
	quit <- struct{}{}
}
