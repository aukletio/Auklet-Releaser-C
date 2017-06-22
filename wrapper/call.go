package main

import "fmt"

type Event struct {
	Fn, Cs uint64
	Type   int
	Time   int64
}

type Call struct {
	Fn, Cs uint64
	Time   int64
}

var s []Event

func push(e Event) {
	s = append(s, e)
}

func pop() Event {
	max := len(s) - 1
	top := s[max]
	s = s[:max]
	return top
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
			f := pop()
			calls <- Call{
				Fn:   e.Fn,
				Cs:   e.Cs,
				Time: e.Time - f.Time,
			}
		}
	}
}
