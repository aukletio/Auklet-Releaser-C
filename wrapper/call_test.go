package main

import (
	"testing"
	"time"
)

var (
	start = Event{
		Frame: Frame{42, 311},
		Type:  0,
		Time:  time.Now().UnixNano(),
	}

	end = Event{
		Frame: start.Frame,
		Type:  1,
		Time:  time.Now().UnixNano(),
	}

	expect = Call{
		Stack: []Frame{start.Frame},
		Time:  end.Time - start.Time,
	}
)

func TestPop(t *testing.T) {
	push(start)
	_, stack := pop()
	if stack[0] != start.Frame {
		t.Fatalf("want %v, got %v", start.Frame, stack[0])
	}
}

func TestCall(t *testing.T) {

	events := make(chan Event)
	calls := make(chan Call)

	go call(events, calls)
	events <- start
	events <- end
	close(events)

	c := <-calls

	switch {
	case expect.Time != c.Time:
		t.Fatalf("want %v, got %v", expect.Time, c.Time)
	case expect.Stack[0] != c.Stack[0]:
		t.Fatalf("want %v, got %v", expect.Stack[0], c.Stack[0])
	}
}
