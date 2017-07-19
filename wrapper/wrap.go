package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime/pprof"

	"github.com/Shopify/sarama"
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

	// TODO: Authenticate the command (associate with an existing release)
	// by computing a checksum of the command binary.

	// Open a socket to communicate with the child command.
	server, err := net.Listen("unix", "socket")
	check(err)
	defer server.Close()

	var producer sarama.AsyncProducer
	if network {
		// Try to connect to the backend so we can post profiles to it.
		producer, err = connect()
		check(err)
		defer producer.Close()
	}

	// Launch the profiler pipeline, since we should be able to receive
	// samples from the child command.

	samples := make(chan []StackFrame, 100)
	profiles := make(chan *Profile)
	done := make(chan struct{}, 2)

	// relay() closes the samples channel when the socket is closed. This
	// causes the concurrent pipeline to shutdown (profiles channels is
	// closed, too). Finally, emit() finishes its work and lets main() know
	// when it's done.

	go emit(profiles, producer, !quiet, done)
	go accumulate(samples, profiles)
	go relay(server, samples)

	// run() might end before or after the socket closes; we don't care
	// which order. We wait for everything to shutdown properly before
	// exiting.

	go run(cmd, done)

	<-done
	<-done
}
