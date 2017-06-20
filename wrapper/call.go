package main

type Event struct {
	Fn, Cs uint64
	Type   int
	Time   int64
}

type Call struct {
	Fn, Cs uint64
	Time   int64
}

func push(s *[]Event, e Event) {
	*s = append(*s, e)
}

func pop(s *[]Event) Event {
	max := len(*s) - 1
	top := (*s)[max]
	*s = (*s)[:max]
	return top
}

func call(events chan Event, calls chan Call) int64 {
	s := make([]Event, 16)

	defer close(calls)

	for {
		e, ok := <-events
		if !ok {
			// channel closed, child exited
			return
		}

		switch e.Type {
		case 0:
			push(&s, e)
		case 1:
			f := pop(&s)
			calls <- Call{
				Fn:   e.Fn,
				Cs:   e.Cs,
				Time: f.Time - e.Time,
			}
		default:
			panic("unreached")
		}
	}
}
