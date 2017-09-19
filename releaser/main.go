package main

import (
	"bytes"
	"crypto/sha512"
	"debug/dwarf"
	"debug/elf"
	"encoding/json"
	"flag"
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
	CommitHash  string            `json:"commit_hash,omitempty"`
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
	log.Printf("usage: %v -apikey apikey -appid appid -deploy deployfile -debug debugfile\n", os.Args[0])
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
}

func main() {
	var deployName, debugName, appID, apiKey string
	flag.StringVar(&deployName, "deploy", "", "ELF binary to be deployed")
	flag.StringVar(&debugName, "debug", "", "ELF binary containing debug symbols")
	flag.StringVar(&appID, "appid", "", "App ID under which to create a release")
	flag.StringVar(&apiKey, "apikey", "", "API key")
	flag.Parse()

	if flag.NFlag() != 4 {
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

	rel.git()

	// create a release
	rel.release(deployName)

	// emit
	b, err := json.MarshalIndent(rel, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))

	body := bytes.NewReader(b)

	urls := map[string]string{
		"production": "https://api.auklet.io/v1/releases/",
		"qa":         "https://api-qa.auklet.io/v1/releases/",
		"staging":    "https://api-staging.auklet.io/v1/releases/",
	}

	endpoint := os.Getenv("AUKLET_RELEASE_ENDPOINT")
	if endpoint == "" {
		endpoint = "production"
	}
	url, in := urls[endpoint]
	if !in {
		panic("releaser: unknown endpoint: " + endpoint)
	}

	// Create a client to control request headers.
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		panic(err)
	}
	req.Header = map[string][]string{
		"content-type": {"application/json"},
		"apikey":       {apiKey},
	}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	log.Println("releaser:", resp.Status)
	log.Printf("releaser:\n" +
	           "    appid: %v\n" +
	           "    commithash: %v\n" +
	           "    checksum: %v\n",
	           rel.AppID,
	           rel.CommitHash,
	           rel.DeployHash)
}
