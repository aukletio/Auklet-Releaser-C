package main

import (
	"encoding/json"
	"fmt"

	"github.com/Shopify/sarama"
)

type Frame struct {
	Fn uint64 `json:",omitempty"`
	Cs uint64 `json:",omitempty"`
}

// A program generally has a tree-like structure, where each node in the tree
// represents a certain point in the callgraph. Linear paths through the
// callgraph are called callchains. A Profile is a node in the callgraph which
// contains aggregate yet context-specific data about program execution.
type Profile struct {
	// The Frame at the end of a given callchain is associated with the
	// sample Count for this point in the callgraph.
	Frame
	Count int `json:",omitempty"`

	// Each leaf Frame has a set of callees, representing possible
	// continuations of the callchain.  Map callee is used to simplify the
	// addCall() algorithm but is not marshaled to JSON; instead, a slice
	// Callee is provided that contains the same data in the format
	// preferred by the backend.

	callee  map[Frame]*Profile
	Callees []*Profile `json:",omitempty"`
}

func NewProfile() *Profile {
	p := new(Profile)
	p.callee = make(map[Frame]*Profile)
	return p
}

func (cur *Profile) addSample(s []Frame) {
	switch len(s) {
	case 0:
		// We reached the top of the stack. Time to add profile data.
		cur.Count++
		return
	default:
		// Eat a stack level and continue.
		f := s[0]
		s = s[1:]

		// Allocate a node for the next frame, if need be.
		next, in := cur.callee[f]
		if !in {
			next = NewProfile()
			cur.callee[f] = next
			cur.Callees = append(cur.Callees, next)
			next.Frame = f
		}

		next.addSample(s)
	}
}

func emit(profiles chan *Profile, producer sarama.AsyncProducer, dump bool, done chan struct{}) {
	for {
		p, ok := <-profiles
		if !ok {
			fmt.Println("pipeline: emit shutting down")
			done <- struct{}{}
			return
		}

		payload, err := json.MarshalIndent(*p, "", "    ")
		check(err)

		if producer != nil {
			publish(producer, payload)
		}

		if dump {
			fmt.Println(string(payload))
		}
	}
}
