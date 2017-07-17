package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// Decode samples from the socket and relay them to an Sample channel.
func relay(server net.Listener, samples chan []Frame) {
	defer func() {
		// There is a race condition between socket EOF and
		// child exit. Nevertheless, if the socket closes, there
		// is nothing left for relay() to do.

		fmt.Println("pipeline: socket EOF: initiating shutdown")
		close(samples)
	}()

	conn, err := server.Accept()
	check(err)

	// Each line represents a sample.
	line := bufio.NewScanner(conn)
	for line.Scan() {
		var s []Frame

		// Each space-delimited word in a line represents a Frame.
		frame := bufio.NewScanner(strings.NewReader(line.Text()))
		frame.Split(bufio.ScanWords)
		for frame.Scan() {
			var f Frame
			// Each colon-delimited word in a Frame represents an
			// address.
			addr := strings.Split(frame.Text(), ":")
			_, err := fmt.Sscanf(addr[0], "%x", &f.Fn)
			check(err)
			_, err = fmt.Sscanf(addr[1], "%x", &f.Cs)
			check(err)
			s = append(s, f)
		}
		samples <- s
	}
}
