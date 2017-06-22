package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Profile struct {
	Time time.Time
	Fn   map[string]Function
	name map[uint64]string
}

type Function struct {
	Name   string
	Ncalls int
	Time   int64
}

func NewProfile(syms map[uint64]string) *Profile {
	p := Profile{}
	p.Fn = make(map[string]Function)
	p.name = syms

	return &p
}

func (prof *Profile) reset() {
	prof.Fn = map[string]Function{}
}

func (prof *Profile) addCall(c Call) {
	name, in := prof.name[c.Fn]
	if !in {
		fmt.Println("No symbol for function address", c.Fn)
		panic(c.Fn)
	}

	F, in := prof.Fn[name]
	if !in {
		// first time using this key
		F.Name = name
	}

	F.Ncalls++
	F.Time += c.Time
	prof.Fn[name] = F
}

func emit(p Profile) {
	p.Time = time.Now()
	b, err := json.MarshalIndent(p, "", "    ")
	check(err)
	fmt.Println(string(b))
}
