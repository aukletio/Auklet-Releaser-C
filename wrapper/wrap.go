package main

import (
	"fmt"
	"os"
	"time"
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
	events := make(chan Event, 100)
	calls := make(chan Call, 100)

	p := NewProfile(syms)
	go call(events, calls)
	go relay("socket", events)
	go run(cmd, child)

	tick := time.Tick(10 * time.Second)
	for {
		select {
		case <-tick:
			emit(*p)
			// reset p
		case c, ok := <-calls:
			if !ok {
				//channel closed, child exited.
				emit(*p)
				return
			}
			p.addCall(c)
		case s := <-sigs:
			// for ease of development, sending SIGINT will cause
			// graceful exit
			fmt.Println(s.String())
			close(events)
		case <-child:
			fmt.Println("wrapper: child exited")
			close(events)
		}
	}
}
