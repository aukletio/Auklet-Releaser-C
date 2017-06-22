package main

import "C"
import (
	"encoding/gob"
	"net"
	"time"
	"unsafe"
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
	// connect to the parent
	var err error
	c, err = net.Dial("unix", "socket")
	if err == nil {
		enc = gob.NewEncoder(c)
	}
}

func sendevent(fn, cs unsafe.Pointer, typ int) {
	e := event{
		Fn:   uint64(uintptr(fn)),
		Cs:   uint64(uintptr(cs)),
		Type: typ,
		Time: time.Now().UnixNano(),
	}

	err := enc.Encode(e)
	check(err)
}

//export __cyg_profile_func_enter
func __cyg_profile_func_enter(fn, cs unsafe.Pointer) {
	if enc != nil {
		sendevent(fn, cs, 0)
	}
}

//export __cyg_profile_func_exit
func __cyg_profile_func_exit(fn, cs unsafe.Pointer) {
	if enc != nil {
		sendevent(fn, cs, 1)
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
}
