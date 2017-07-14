package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime/pprof"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func usage() {
	fmt.Printf("usage: %v command [args ...]\n", os.Args[0])
	os.Exit(0)
}

func main() {
	var cpuprofile, network, quiet bool

	flag.BoolVar(&cpuprofile, "p", false, "compute wrapper cpu profile")
	flag.BoolVar(&network, "n", false, "publish profiles to backend")
	flag.BoolVar(&quiet, "q", false, "do not dump profiles to stdout")

	flag.Parse()

	if cpuprofile {
		f, err := os.Create("cpuprofile")
		check(err)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	args := flag.Args()
	if len(args) < 1 {
		usage()
	}
	cmd := exec.Command(args[0], args[1:]...)

	// TODO: Authenticate the command (associate with an existing release).
	// If via HTTPS, do it here; if via MQTT, do it after calling
	// connect(). Don't profile if not associated with a release.

	// Open a socket to communicate with the child command.
	server, err := net.Listen("unix", "socket")
	check(err)
	defer server.Close()

	var client mqtt.Client
	if network {
		// Try to connect to the backend so we can post profiles to it.
		client, err = connect()
		check(err)
		defer client.Disconnect(250)
	}

	// Launch the profiler pipeline, since we should be able to receive
	// samples from the child command.

	samples := make(chan []Frame, 100)
	profiles := make(chan *Profile)
	done := make(chan struct{}, 2)

	// relay() closes the samples channel when the socket is closed. This
	// causes the concurrent pipeline to shutdown (profiles channels is
	// closed, too). Finally, emit() finishes its work and lets main() know
	// when it's done.

	go emit(profiles, client, !quiet, done)
	go accumulate(samples, profiles)
	go relay(server, samples)

	// run() might end before or after the socket closes; we don't care
	// which order. We wait for everything to shutdown properly before
	// exiting.

	go run(cmd, done)

	<-done
	<-done
}
