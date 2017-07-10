package main

import (
	"fmt"
)

// The functions __cyg_profile_func_{enter,exit} within the instrument take
// Frames as arguments; Frames consist of a function address (a function
// pointer) and a callsite address (the address of the call instruction).
type Frame struct {
	Fn, Cs uint64
}

// The instrument emits start (type 0) and end (type 1) events over the socket
// with a timestamp in Unix nanoseconds.
type Event struct {
	Frame
	Type int
	Time int64
}

// A call is a completed function call; it contains a snapshot of the stack
// during its execution, which is used to create a callgraph.
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
			// The socket gave an EOF. No more events will be
			// generated; thus no more calls can be added to a
			// profile.

			return
		}

		if e.Type == 0 {
			push(e)
		} else {
			time, stack := pop()

			// TODO: Compute the call duration in a way that
			// excludes the time spent in callees.

			calls <- Call{
				Stack: stack,
				Time:  e.Time - time,
			}
		}
	}
}
