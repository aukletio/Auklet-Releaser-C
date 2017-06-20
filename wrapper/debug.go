package main

import (
	"debug/dwarf"
	"debug/elf"
	"fmt"
	"io"
	"os"
)

func symbols(path string) map[uint64]string {
	f, err := elf.Open(path)
	check(err)

	defer func() {
		err := f.Close()
		check(err)
	}()

	symbols, err := f.Symbols()
	if err != nil {
		fmt.Printf("error: file %v has no symbol section\n", path)
		os.Exit(1)
	}

	s := make(map[uint64]string)
	for _, sym := range symbols {
		if sym.Value == 0 || sym.Name == "" {
			continue
		}

		s[sym.Value] = sym.Name
	}

	return s
}

func addrMap(path string) map[uint64]dwarf.LineEntry {
	f, err := elf.Open(path)
	check(err)

	defer func() {
		err := f.Close()
		check(err)
	}()

	d, err := f.DWARF()
	check(err)
	r := d.Reader()

	dwarfLineEntry := make(map[uint64]dwarf.LineEntry)
	for {
		entry, err := r.Next()
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

	return dwarfLineEntry
}
