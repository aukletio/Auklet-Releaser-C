package main

import (
	"encoding/json"
	"fmt"
)

type Profile struct {
	Frame
	Ncalls  int
	Time    int64
	callee  map[Frame]*Profile
	Callees []*Profile
}

func NewProfile() *Profile {
	p := new(Profile)
	p.callee = make(map[Frame]*Profile)
	//p.Callees = make([]*Profile, 0)
	return p
}

func (cur *Profile) addCall(c Call) {
	switch len(c.Stack) {
	case 0:
		// We reached the leaf. Time to add profile data.
		cur.Ncalls++
		cur.Time += c.Time
		return
	default:
		// Eat a stack level and continue.
		f := c.Stack[0]
		c.Stack = c.Stack[1:]

		// Allocate a node for the next frame, if need be.
		next, in := cur.callee[f]
		if !in {
			next = NewProfile()
			cur.callee[f] = next
			cur.Callees = append(cur.Callees, next)
			next.Frame = f
		}

		next.addCall(c)
	}
}

func emit(p *Profile) {
	b, err := json.MarshalIndent(*p, "", "    ")
	check(err)
	fmt.Println(string(b))
}
