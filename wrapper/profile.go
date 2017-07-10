package main

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// A program generally has a tree-like structure, where each node in the tree
// represents a certain point in the callgraph. Linear paths through the
// callgraph are called callchains. A Profile is a node in the callgraph which
// contains aggregate yet context-specific data about program execution.
type Profile struct {
	// The Frame at the end of a given callchain is associated with the
	// number of calls and total time spent at this point in the callgraph.
	Frame
	Ncalls int   `json:",omitempty"`
	Time   int64 `json:",omitempty"`

	// This map is used to simplify the addCall() algorithm. A Profile is
	// marshaled to JSON and thus a Go map[Frame] is unsuitable.
	callee map[Frame]*Profile

	// Each leaf Frame has callees, representing possible continuations of
	// the callchain.
	Callees []*Profile `json:",omitempty"`
}

func NewProfile() *Profile {
	p := new(Profile)
	p.callee = make(map[Frame]*Profile)
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

func emit(profiles chan *Profile, c mqtt.Client, dump bool, done chan struct{}) {
	for {
		p, ok := <-profiles
		if !ok {
			fmt.Println("pipeline: emit shutting down")
			done <- struct{}{}
			return
		}

		payload, err := json.MarshalIndent(*p, "", "    ")
		check(err)

		if c != nil {
			publish(c, payload)
		}

		if dump {
			fmt.Println(string(payload))
		}
	}
}
