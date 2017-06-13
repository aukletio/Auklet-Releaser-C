package main

import (
	"C"
	"debug/dwarf"
	"debug/elf"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"time"
	"unsafe"
)

type Event struct {
	Faddr, Caddr uint64
	Time         int64
}

type Profile struct {
	Time time.Time
	Sig  string
	// key is stringified callsite address
	Cs map[string]Callsite
}

type Callsite struct {
	Name     string
	Ncalls   int
	Time     int64
	At, Defn dwarf.LineEntry
}

var (
	e              []Event
	calls          chan Event
	quit           chan int
	profiles       chan Profile
	prof           Profile
	elfSymbol      map[uint64]elf.Symbol
	dwarfLineEntry map[uint64]dwarf.LineEntry
	ppid           int
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// pre-main constructor
func init() {
	// In theory, this prevents context switching in the main thread.
	runtime.LockOSThread()
	ppid = os.Getpid()
	readDebug()
	quit = make(chan int, 0)
	calls = make(chan Event, 0)
	sigs := make(chan os.Signal)
	signal.Notify(sigs)
	go func() {
		for s := range sigs {
			prof.Time = time.Now()
			prof.Sig = s.String()
			emit(prof)
			os.Exit(0)
		}
	}()
	//profiles = make(chan Profile, 0)
	go profile()
	//go emit()
}

func readDebug() {
	f, err := elf.Open(os.Args[0])
	check(err)

	symbols, err := f.Symbols()
	check(err)

	elfSymbol = make(map[uint64]elf.Symbol)
	// scan ELF symbol table to associate function names with addresses
	for _, sym := range symbols {
		// assume zero values are dynamically linked
		if sym.Value == 0 {
			continue
		}

		elfSymbol[sym.Value] = sym
	}

	d, err := f.DWARF()
	check(err)
	r := d.Reader()

	dwarfLineEntry = make(map[uint64]dwarf.LineEntry)
	// Scan DWARF info to associate functions with source Locations.
	for {
		entry, err := r.Next()
		check(err)
		// skip uninteresting debug entries
		if entry == nil {
			break
		}
		if entry.Tag != dwarf.TagCompileUnit {
			continue
		}
		lr, err := d.LineReader(entry)
		check(err)
		if lr == nil {
			continue
		}

		// inspect line entries
		for {
			var le dwarf.LineEntry
			err := lr.Next(&le)
			if err == io.EOF {
				break
			}
			check(err)
			dwarfLineEntry[le.Address] = le
		}
	}
	err = f.Close()
	check(err)
}

func profile() {
	tick := time.Tick(10 * time.Second)
	var cs string
	prof.Cs = make(map[string]Callsite)
	defer func() {
		fmt.Println("profile: exiting")
		if x := recover(); x != nil {
			fmt.Printf("run time panic: %v", x)
		}
	}()
	for {
		select {
		case <-quit:
			prof.Time = time.Now()
			emit(prof)
			quit <- 0
		case t := <-tick:
			prof.Time = t
			emit(prof)
			//profiles <- prof

			// reset the profile
			prof.Cs = map[string]Callsite{}
		case c := <-calls:
			cs = fmt.Sprint(c.Caddr)
			K, ok := prof.Cs[cs]
			if !ok {
				K.At = dwarfLineEntry[c.Caddr]
				K.Name = elfSymbol[c.Faddr].Name
				K.Defn = dwarfLineEntry[c.Faddr]
			}
			K.Ncalls++
			K.Time += c.Time
			prof.Cs[cs] = K
		}
	}
}

func emit(p Profile) {
	b, err := json.MarshalIndent(p, "", "        ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func push(val Event) {
	if ppid != os.Getpid() {
		return
	}
	e = append(e, val)
}

func pop() {
	defer func() {
		//fmt.Println("pop: exiting")
		if x := recover(); x != nil {
			fmt.Printf("run time panic: %v", x)
		}
	}()
	var result Event
	var i int
	if ppid != os.Getpid() {
		return
	}
	i = len(e) - 1
	result = e[i]
	e = e[:i]
	result.Time += time.Now().UnixNano()
	if i != 0 {
		e[i-1].Time -= result.Time
	} else {
		calls <- result
		quit <- 0
		<-quit
		return
	}
	calls <- result
}

//export __cyg_profile_func_enter
func __cyg_profile_func_enter(faddr, caddr unsafe.Pointer) {
	push(Event{
		uint64(uintptr(faddr)),
		uint64(uintptr(caddr)),
		-time.Now().UnixNano(),
	})
}

//export __cyg_profile_func_exit
func __cyg_profile_func_exit(faddr, caddr unsafe.Pointer) {
	pop()
}

func main() {
}
