package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"
)

func main() {
	f, err := os.Create("cpuprofile")
	check(err)
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	// handle signals
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT)

	// define a command
	cmd := command()

	// connect to server
	client, err := connect()
	check(err)
	defer client.Disconnect(250)

	child := make(chan struct{})
	events := make(chan Event, 1000)
	calls := make(chan Call, 1000)

	p := NewProfile()
	defer emit(client, p)

	go call(events, calls)
	server, err := net.Listen("unix", "socket")
	defer func() {
		if server != nil {
			server.Close()
		}
	}()
	check(err)

	go relay(server, events)
	go run(cmd, child)

	tick := time.Tick(1 * time.Minute)
	for {
		select {
		case <-tick:
			emit(client, p)
			p = NewProfile()
		case c, ok := <-calls:
			if !ok {
				//channel closed, child exited.
				return
			}
			p.addCall(c)
		case s := <-sigs:
			// for ease of development, sending SIGINT will cause
			// graceful exit
			fmt.Println(s.String())
			return
		case <-child:
			fmt.Println("wrapper: child exited")
			return
		}
	}
}
