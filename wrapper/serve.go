package main

import (
	"encoding/gob"
	"io"
	"net"
)

// Decode events from the socket and relay them to an Event channel.
func relay(server net.Listener, events chan Event) {
	defer close(events)

	conn, err := server.Accept()
	check(err)

	dec := gob.NewDecoder(conn)

	for {
		var e Event
		err := dec.Decode(&e)
		if err == io.EOF {

			// There is a race condition between socket EOF and
			// child exit. Nevertheless, if the socket closes, there
			// is nothing left for relay() to do.

			return
		}
		events <- e
	}
}
