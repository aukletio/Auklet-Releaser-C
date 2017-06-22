package main

import (
	"encoding/gob"
	"io"
	"net"
)

func relay(server net.Listener, events chan Event) {
	defer close(events)

	conn, err := server.Accept()
	check(err)

	dec := gob.NewDecoder(conn)

	for {
		var e Event
		err := dec.Decode(&e)
		if err == io.EOF {
			return
		}
		events <- e
	}
}
