package main

import (
	"fmt"
)

type Frame struct {
	Fn, Cs uint64
}

type Event struct {
	Frame
	Type int
	Time int64
}

type Call struct {
	Stack []Frame
	Time  int64
}

var fs []Frame
var ts []int64

func push(e Event) {
	fs = append(fs, e.Frame)
	ts = append(ts, e.Time)
}

func pop() (int64, []Frame) {
	f := make([]Frame, len(fs))
	max := len(ts) - 1

	// Pop from the time stack.
	time := ts[max]
	ts = ts[:max]

	// Pop from the frame stack.
	copy(f, fs)
	fs = fs[:max]
	return time, f
}

func call(events chan Event, calls chan Call) {
	defer func() {
		if x := recover(); x != nil {
			fmt.Println("wrapper: call:", x)
		}

		close(calls)
	}()

	for {
		e, ok := <-events
		if !ok {
			// channel closed, child exited
			return
		}

		if e.Type == 0 {
			push(e)
		} else {
			time, stack := pop()
			calls <- Call{
				Stack: stack,
				Time:  e.Time - time,
			}
		}
	}
}
