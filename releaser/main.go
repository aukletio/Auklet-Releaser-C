package main

import (
	"bytes"
	"crypto/md5"
	"debug/dwarf"
	"debug/elf"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

type Release struct {
	DeployHash string            `json:"checksum"`
	Symbols    []elf.Symbol      `json:"symbols"`
	Dwarf      []dwarf.LineEntry `json:"dwarf"`
	AppID string `json:"app_id"`
}

func usage() {
	fmt.Printf("usage: %v -appid appid -deploy deployfile -debug debugfile\n", os.Args[0])
	os.Exit(1)
}

func (rel *Release) symbolize(debugpath string) {
	debugfile, err := elf.Open(debugpath)
	if err != nil {
		log.Fatal(err)
	}
	defer debugfile.Close()

	// add ELF symbols
	rel.Symbols, err = debugfile.Symbols()
	if err != nil {
		log.Fatal(err)
	}

	// add DWARF line entries
	d, err := debugfile.DWARF()
	if err != nil {
		log.Fatal(err)
	}
	r := d.Reader()

	for {
		entry, err := r.Next()
		if entry == nil {
			break
		}
		if entry.Tag != dwarf.TagCompileUnit {
			continue
		}
		lr, err := d.LineReader(entry)
		if err != nil {
			log.Fatal(err)
		}
		if lr == nil {
			continue
		}

		for {
			var le dwarf.LineEntry
			err := lr.Next(&le)
			if err == io.EOF {
				break
			} else if err != nil {
				panic(err)
			}
			rel.Dwarf = append(rel.Dwarf, le)
		}
	}
}

func hash(s *elf.Section) []byte {
	r := s.Open()
	h := md5.New()
	if _, err := io.Copy(h, r); err != nil {
		log.Fatal(err)
	}
	return h.Sum(nil)
}

func sectionsMatch(deployName, debugName string) bool {
	deployfile, err := elf.Open(deployName)
	if err != nil {
		log.Fatal(err)
	}
	defer deployfile.Close()

	debugfile, err := elf.Open(deployName)
	if err != nil {
		log.Fatal(err)
	}
	defer debugfile.Close()

	// compare file sections
	for _, deploysect := range deployfile.Sections {
		if deploysect == nil || deploysect.Type == elf.SHT_STRTAB {
			continue
		}

		debugsect := debugfile.Section(deploysect.Name)
		if debugsect == nil {
			fmt.Printf("debug file %v lacks section %v from deploy file %v\n",
				debugName, deploysect.Name, deployName)
			continue
		}

		if bytes.Compare(hash(deploysect), hash(debugsect)) != 0 {
			fmt.Printf("section %-15v %-18v differs\n",
				deploysect.Name, deploysect.Type)
			return false
		}
	}
	return true
}

func (rel *Release) release(deployName string) {
	f, err := os.Open(deployName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	dh := md5.New()
	if _, err := io.Copy(dh, f); err != nil {
		log.Fatal(err)
	}
	deployhash := dh.Sum(nil)
	rel.DeployHash = fmt.Sprintf("%x", deployhash)
}

func main() {
	var deployName, debugName, appID string
	flag.StringVar(&deployName, "deploy", "", "ELF binary to be deployed")
	flag.StringVar(&debugName, "debug", "", "ELF binary containing debug symbols")
	flag.StringVar(&appID, "appid", "", "App ID under which to create a release")
	flag.Parse()

	if flag.NFlag() != 3 {
		usage()
	}

	rel := new(Release)
	rel.AppID = appID

	// get debug info and symbols
	rel.symbolize(debugName)

	// reject ELF pairs with disparate sections
	if !sectionsMatch(deployName, debugName) {
		os.Exit(1)
	}

	// create a release
	rel.release(deployName)

	// emit
	b, err := json.MarshalIndent(rel, "", "    ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}
