package main

import (
	"fmt"
	"time"
)

func accumulate(samples chan []Frame, profiles chan *Profile) {
	p := NewProfile()

	defer close(profiles)
	tick := time.Tick(1 * time.Minute)
	for {
		select {
		case s, ok := <-samples:
			if !ok {
				profiles <- p
				fmt.Println("pipeline: accumulate shutting down")
				return
			}
			p.addSample(s)
		case <-tick:
			profiles <- p
			p = NewProfile()
		}
	}
}
