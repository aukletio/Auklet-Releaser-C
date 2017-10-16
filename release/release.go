package main

import (
	"bufio"
	"bytes"
	"crypto/sha512"
	"debug/dwarf"
	"debug/elf"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

// A Release represents a release of a customer's app to be sent to the backend.
type Release struct {
	AppID       string            `json:"app_id"`
	DeployHash  string            `json:"checksum"`
	CommitHash  string            `json:"scm_data,omitempty"`
	GitTopLevel string            `json:"absolute_path_prefix,omitempty"`
	Dwarf       []dwarf.LineEntry `json:"dwarf"`
	Symbols     []elf.Symbol      `json:"symbols"`
}

// A BytesReadCloser is a bytes.Reader that satisfies io.ReadCloser, which is
// necessary for it to be used in an http.Request.
type BytesReadCloser bytes.Reader

func (s BytesReadCloser) Close() error {
	return nil
}

func usage() {
	log.Printf("usage: %v deployfile\n", os.Args[0])
	os.Exit(1)
}

func (rel *Release) symbolize(debugpath string) {
	debugfile, err := elf.Open(debugpath)
	if err != nil {
		log.Println("Debug filename must be <deployfile>-dbg")
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

func (rel *Release) git() error {
	// Associate a release with the top-level directory of the Git repo.
	gtl := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := gtl.CombinedOutput()
	if err != nil {
		// Not a git repo or don't have git
		log.Println(string(out))
		panic(err)
	}

	rel.GitTopLevel = string(out[:len(out)-1])

	// Associate the release with the current commit hash.
	rph := exec.Command("git", "rev-parse", "HEAD")
	out, err = rph.Output()
	if err != nil {
		log.Println(string(out))
		panic(err)
	}

	rel.CommitHash = string(out[:len(out)-1])
	return nil
}

func hash(s *elf.Section) []byte {
	r := s.Open()
	h := sha512.New512_224()
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
			log.Printf("releaser: debug file %v lacks section %v from deploy file %v\n",
				debugName, deploysect.Name, deployName)
			continue
		}

		if bytes.Compare(hash(deploysect), hash(debugsect)) != 0 {
			log.Printf("releaser: section %-15v %-18v differs\n",
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

	dh := sha512.New512_224()
	if _, err := io.Copy(dh, f); err != nil {
		log.Fatal(err)
	}
	deployhash := dh.Sum(nil)
	rel.DeployHash = fmt.Sprintf("%x", deployhash)
	log.Println("release():", deployName, rel.DeployHash)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	deployName := os.Args[1]
	debugName := deployName + "-dbg"

	url := func() string {
		endpoint := os.Getenv("AUKLET_ENDPOINT")
		if endpoint == "" {
			log.Fatal("empty envar AUKLET_ENDPOINT")
		}
		f, err := os.Open(endpoint + "/url")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		s, err := bufio.NewReader(f).ReadString('\n')
		if err != nil {
			panic(err)
		}
		return s[:len(s)-1]
	}() + "/releases/"
	log.Println("releaser: url:", url)

	rel := new(Release)
	rel.AppID = os.Getenv("AUKLET_APP_ID")
	if rel.AppID == "" {
		log.Fatal("empty envar AUKLET_APP_ID")
	}
	apikey := os.Getenv("AUKLET_API_KEY")
	if rel.AppID == "" {
		log.Fatal("empty envar AUKLET_API_KEY")
	}
	rel.symbolize(debugName)

	// reject ELF pairs with disparate sections
	if !sectionsMatch(deployName, debugName) {
		os.Exit(1)
	}

	rel.git()

	// create a release
	rel.release(deployName)

	// emit
	b, err := json.MarshalIndent(rel, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))

	// Create a client to control request headers.
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		panic(err)
	}
	req.Header = map[string][]string{
		"content-type": {"application/json"},
		"apikey":       {apikey},
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	log.Println("releaser:", resp.Status)
	switch resp.StatusCode {
	case 200:
		log.Println("not created")
	}
	log.Printf("releaser:\n"+
		"    appid: %v\n"+
		"    commithash: %v\n"+
		"    checksum: %v\n",
		rel.AppID,
		rel.CommitHash,
		rel.DeployHash)
}
