package main

import (
	"time"
)

func accumulate(samples chan []StackFrame, profiles chan *Profile) {
	p := NewProfile()

	defer close(profiles)
	tick := time.Tick(1 * time.Minute)
	for {
		select {
		case s, ok := <-samples:
			if !ok {
				profiles <- p
				return
			}
			p.addSample(s)
		case <-tick:
			profiles <- p
			p = NewProfile()
		}
	}
}
