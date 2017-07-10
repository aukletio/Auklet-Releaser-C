package main

import (
	"net"
	"fmt"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
)

func main() {
	// Handle SIGINT so that ^C causes a clean exit.
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT)

	// TODO: Use a command-line flag to start CPU profiling.
	f, err := os.Create("cpuprofile")
	check(err)
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	cmd := command()

	var done chan struct{}

	// TODO: Authenticate the command (associate with an existing release).
	// If via HTTPS, do it here; if via MQTT, do it after calling
	// connect(). Don't profile if not associated with a release.

	// Try to connect to the backend so we can post profiles to it.
	client, err := connect()
	if err != nil {
		defer client.Disconnect(250)

		// Open a socket to communicate with the child command.
		server, err := net.Listen("unix", "socket")
		if err != nil {
			defer server.Close()

			// Launch the profiler pipeline, since we should be able to receive
			// events from the child command and post profiles to the backend.

			events := make(chan Event, 1000)
			calls := make(chan Call, 1000)

			go call(events, calls)
			go relay(server, events)
			go accumulate(calls, client, sigs, done)
		} else {
			fmt.Println("no socket")
			// Continue even without the socket. We should not prevent a
			// program from being run just because we can't profile it.

			// TODO: Consider telling the backend that the socket connection
			// failed. That's odd.

		}
	} else {
		fmt.Println("no backend")

		// Continue even without a backend connection. We should not
		// prevent a program from being run just because we failed
		// to connect to the backend.

	}

	done = make(chan struct{})
	run(cmd)
}
