package main

import (
	"fmt"
	"time"
	"os"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func accumulate(calls chan Call, client mqtt.Client, sigs <-chan os.Signal, done chan struct{}) {
	p := NewProfile()

	tick := time.Tick(1 * time.Minute)
	for {
		select {
		case c, ok := <-calls:
			if !ok {

				// The socket gave an EOF. We post any
				// accumulated profile data; there is no
				// more work to do.

				// TODO: Tell the backend that the
				// socket was closed.  There seems to be
				// a race condition between child exit
				// and socket close.

				emit(client, p)
				fmt.Println("wrapper: socket closed")
				return
			}
			p.addCall(c)
		case <-tick:
			emit(client, p)
			p = NewProfile()
		case s := <-sigs:
			fmt.Println(s.String())
			return
		case <-done:

			// TODO: Tell the backend that the child
			// exited; report its exit status.

			emit(client, p)
			fmt.Println("wrapper: child exited")
			return
		}
	}
}
