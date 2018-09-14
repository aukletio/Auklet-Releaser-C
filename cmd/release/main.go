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
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gobuffalo/packr"

	"github.com/ESG-USA/Auklet-Releaser-C/config"
)

// A Dwarf represents a pared-down dwarf.LineEntry.
type Dwarf struct {
	Address  uint64
	FileName string
	Line     int
}

// A Symbol represents a pared-down elf.Symbol.
type Symbol struct {
	Name  string
	Value uint64
}

type languageMeta struct {
	TopLevel   string   `json:"absolute_path_prefix"`
	DeployHash string   `json:"checksum"`
	Dwarf      []Dwarf  `json:"dwarf"`
	Symbols    []Symbol `json:"symbols"`
}

// A Release represents a release of a customer's app to be sent to the backend.
type Release struct {
	AppID        string `json:"application"`
	languageMeta `json:"language_meta"`
	CommitHash   string `json:"commit_hash"`
}

func usage() {
	fmt.Printf("usage: %v deployfile\n", os.Args[0])
	fmt.Printf("view OSS licenses: %v --licenses\n", os.Args[0])
}

func licenses() {
	licensesBox := packr.NewBox("./licenses")
	licenses := licensesBox.List()
	// Print the Auklet license first, then iterate over all the others.
	format := "License for %v\n-------------------------\n%v"
	fmt.Printf(format, "Auklet Releaser", licensesBox.String("LICENSE"))
	for _, l := range licenses {
		if l != "LICENSE" {
			ownerName := strings.Split(l, "--")
			fmt.Printf("\n\n\n")
			header := fmt.Sprintf("package: %v/%v", ownerName[0], ownerName[1])
			fmt.Printf(format, header, licensesBox.String(l))
		}
	}
}

func (rel *Release) symbolize(debugpath string) {
	debugfile, err := elf.Open(debugpath)
	if err != nil {
		log.Println("Debug filename must be <deployfile>-dbg")
		log.Fatal(err)
	}
	defer debugfile.Close()

	// add ELF symbols
	ss, err := debugfile.Symbols()
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range ss {
		rel.Symbols = append(rel.Symbols, Symbol{
			Name:  s.Name,
			Value: s.Value,
		})
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
			if le.File == nil {
				continue
			}
			_, err = os.Stat(le.File.Name)
			if err == nil {
				rel.Dwarf = append(rel.Dwarf, Dwarf{
					Address:  le.Address,
					FileName: le.File.Name,
					Line:     le.Line,
				})
			}
		}
	}
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
			log.Printf("debug file %v lacks section %v from deploy file %v\n",
				debugName, deploysect.Name, deployName)
			continue
		}

		if bytes.Compare(hash(deploysect), hash(debugsect)) != 0 {
			log.Printf("section %-15v %-18v differs\n",
				deploysect.Name, deploysect.Type)
			return false
		}
	}
	return true
}

func (rel *Release) commitHash() {
	c := exec.Command("git", "rev-parse", "HEAD")
	out, err := c.CombinedOutput()
	if err != nil {
		log.Print(err)
	} else {
		rel.CommitHash = strings.TrimSpace(string(out))
	}
}

func (rel *Release) topLevel() {
	c := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := c.CombinedOutput()
	if err != nil {
		log.Print(err)
	} else {
		rel.TopLevel = strings.TrimSpace(string(out))
	}
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

func getConfig() config.Config {
	var cfg config.Config
	if Version == "local-build" {
		cfg = config.LocalBuild()
	} else {
		cfg = config.ReleaseBuild()
	}
	if !cfg.Valid() {
		log.Fatal("incomplete configuration")
	}
	return cfg
}

func newRelease(deployName, appID string) *Release {
	rel := new(Release)
	rel.AppID = appID
	debugName := deployName + "-dbg"
	rel.symbolize(debugName)

	// reject ELF pairs with disparate sections
	if !sectionsMatch(deployName, debugName) {
		os.Exit(1)
	}

	rel.commitHash()
	rel.topLevel()
	rel.release(deployName)
	return rel
}

func post(rel *Release, cfg config.Config) {
	b, err := json.MarshalIndent(rel, "", "    ")
	if err != nil {
		panic(err)
	}

	url := cfg.BaseURL + "/v1/releases/"
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		panic(err)
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", "JWT "+cfg.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	log.Printf("appid: %v\n", rel.AppID)
	log.Printf("checksum: %v\n", rel.DeployHash)
	log.Print(resp.Status)
	switch resp.StatusCode {
	case 200:
		log.Println("not created")
	case 201: // created
	case 502: // bad gateway
		log.Fatal(url)
	default:
		b, _ := ioutil.ReadAll(resp.Body)
		log.Print(string(b))
	}
}

var (
	viewLicenses bool
	version string
	release string
)

func init() {
	flag.BoolVar(&viewLicenses, "licenses", false, "view OSS licenses")
	flag.StringVar(&release, "release", "", "override Git commit hash with user-defined release identifier")
	flag.StringVar(&version, "version", "", "user-defined version string")
}

func main() {
	flag.Parse()
	if viewLicenses {
		licenses()
		os.Exit(1)
	}

	args := flag.Args()
	if len(flag.Args()) == 0 {
		usage()
		os.Exit(1)
	}

	log.Printf("Auklet Releaser version %s (%s)\n", Version, BuildDate)

	cfg := getConfig()
	rel := newRelease(args[0], cfg.AppID)
	post(rel, cfg)
}
