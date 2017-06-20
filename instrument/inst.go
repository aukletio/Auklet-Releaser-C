package main

import "C"
import (
	"encoding/gob"
	"net"
	"time"
	"unsafe"
)

const (
	start = iota
	end   = iota
)

type event struct {
	Fn, Cs uint64
	Typ    int
	Time   int64
}

var (
	c   net.Conn
	enc *gob.Encoder
)

func init() {
	var err error

	// connect to the parent
	c, err = net.Dial("unix", "profiler.sock")
	check(err)

	enc = gob.NewEncoder(c)
}

func sendevent(fn, cs unsafe.Pointer, typ int) {
	check(enc.Encode(event{
		Fn:   uint64(uintptr(fn)),
		Cs:   uint64(uintptr(cs)),
		Typ:  typ,
		Time: time.Now().UnixNano(),
	}))
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
