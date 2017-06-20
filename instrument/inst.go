package main

import "C"
import (
	"encoding/gob"
	"net"
	"fmt"
	"time"
	"unsafe"
)

const (
	start = 0
	end   = 1
)

type event struct {
	Fn, Cs uint64
	Type   int
	Time   int64
}

var (
	c   net.Conn
	enc *gob.Encoder
)

func init() {
	var err error

	// connect to the parent
	c, err = net.Dial("unix", "socket")
	check(err)

	enc = gob.NewEncoder(c)
}

func sendevent(fn, cs unsafe.Pointer, typ int) {
	e := event{
		Fn:   uint64(uintptr(fn)),
		Cs:   uint64(uintptr(cs)),
		Type: typ,
		Time: time.Now().UnixNano(),
	}
	fmt.Println("instrument:", e)
	err := enc.Encode(e)
	check(err)
}

//export __cyg_profile_func_enter
func __cyg_profile_func_enter(fn, cs unsafe.Pointer) {
	sendevent(fn, cs, start)
}

//export __cyg_profile_func_exit
func __cyg_profile_func_exit(fn, cs unsafe.Pointer) {
	sendevent(fn, cs, end)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
}
