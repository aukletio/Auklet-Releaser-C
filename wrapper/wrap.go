package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// handle signals
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT)

	// define a command
	cmd := command()

	// read ELF symbols
	syms := symbols(cmd.Path)

	child := make(chan struct{})
	events := make(chan Event)
	calls := make(chan Call)
	profiles := make(chan Profile)

	p := NewProfile(syms)
	go emit(profiles)
	go call(events, calls)
	go relay("profiler.sock", events)
	go run(cmd, child)

	for {
		select {
		case <-tick:
			emit(p)
			// reset p
		case c, ok := <-calls:
			if !ok {
				//channel closed, child exited.
			}
			p.addCall(c)
		case s := <-sigs:
			fmt.Println(s.String())
			emit(p)
			return
		case <-child:
			close(events)
			emit(p)
			return
		}
	}
}
