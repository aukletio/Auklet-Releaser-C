package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Profile struct {
	Time time.Time
	Root *Node
	name map[uint64]string
}

type Node struct {
	Ncalls int
	Time   int64
	Callee map[Frame]*Node
}

func (cur *Node) addCall(c Call) {
	switch len(c.Stack) {
	case 0:
		// Empty stack, something is wrong.
		panic("unreached")
	case 1:
		// We reached the leaf. Time to add profile data.
		cur.Ncalls++
		cur.Time += c.Time
		return
	default:
		// Eat the next stack level and continue.
		f := c.Stack[0]
		c.Stack = c.Stack[1:]

		// Allocate a map, if need be.
		if cur.Callee == nil {
			cur.Callee = make(map[Frame]*Node)
		}

		// Allocate a node for this frame, if need be.
		next, in := cur.Callee[f]
		if !in {
			next = new(Node)
			cur.Callee[f] = next
		}

		next.addCall(c)
	}
}

func emit(p Profile) {
	p.Time = time.Now()
	b, err := json.MarshalIndent(p, "", "    ")
	check(err)
	fmt.Println(string(b))
}

func NewProfile(syms map[uint64]string) *Profile {
	p := Profile{}
	p.Root = new(Node)
	p.name = syms

	return &p
}

