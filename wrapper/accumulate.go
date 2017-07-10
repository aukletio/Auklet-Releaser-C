package main

import (
	"fmt"
	"time"
)

func accumulate(calls chan Call, profiles chan *Profile) {
	p := NewProfile()

	defer close(profiles)
	tick := time.Tick(1 * time.Minute)
	for {
		select {
		case c, ok := <-calls:
			if !ok {
				profiles <- p
				fmt.Println("pipeline: accumulate shutting down")
				return
			}
			p.addCall(c)
		case <-tick:
			profiles <- p
			p = NewProfile()
		}
	}
}
