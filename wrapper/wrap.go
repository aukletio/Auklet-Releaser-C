package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime/pprof"
	"strconv"

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
	sigs := make(chan os.Signal)
	signal.Notify(sigs)

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
	server, err := net.Listen("unix", "socket-"+strconv.Itoa(os.Getpid()))
	check(err)
	defer server.Close()

	var producer sarama.SyncProducer
	if network {
		// Try to connect to the backend so we can post profiles to it.
		producer, err = connect()
		check(err)
		defer producer.Close()
	}

	done := make(chan struct{}, 2)
	go func() {
		for s := range sigs {
			fmt.Println(s)
			server.Close()
		}
	}()

	go relay(server, producer, done)
	go run(cmd, done)
	<-done
	<-done
}
