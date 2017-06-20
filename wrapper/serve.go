package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"net"
)

func relay(addr string, events chan Event) {
	var server net.Listener
	defer func() {
		if x := recover(); x != nil {
			fmt.Println(x)
		}

		server.Close()
	}()

	server, err := net.Listen("unix", addr)
	check(err)

	conn, err := server.Accept()
	check(err)

	var e Event

	dec := gob.NewDecoder(conn)

	for err != io.EOF {
		err = dec.Decode(&e)
		events <- e
	}
}
