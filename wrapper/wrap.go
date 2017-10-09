package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime/pprof"
	"strconv"
	"sync"

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

	cksum := checksum(cmd.Path)
	if !valid(cksum) {
		log.Fatal("wrapper: invalid checksum")
	}

	// Open a socket to communicate with the child command.
	server, err := net.Listen("unix", "socket-"+strconv.Itoa(os.Getpid()))
	check(err)
	defer server.Close()

	go func() {
		for s := range sigs {
			log.Println(s)
			server.Close()
			os.Exit(0)
		}
	}()

	msg := make(chan sarama.ProducerMessage)
	go produce(msg)

	var wg sync.WaitGroup
	if network {
		wg.Add(1)
		go relay(server, msg, cksum, &wg)
	}

	wg.Add(1)
	go run(cmd, msg, &wg)
	wg.Wait()
}
