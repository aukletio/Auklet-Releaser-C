package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/sha512"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/satori/go.uuid"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	hnet "github.com/shirou/gopsutil/net"
)

// BuildDate is provided at compile-time; DO NOT MODIFY.
var BuildDate = "no timestamp"

// Version is provided at compile-time; DO NOT MODIFY.
var Version = "local-build"

// Object represents something that can be sent to the backend. It must have a
// topic and implement a brand() method that fills UUID and checksum fields.
type Object interface {
	topic() string
	brand()
}

func checksum(path string) string {
	f, err := os.Open(path)
	if err != nil {
		log.Panic(err)
	}
	defer f.Close()

	h := sha512.New512_224()
	if _, err := io.Copy(h, f); err != nil {
		log.Panic(err)
	}

	hash := h.Sum(nil)
	sum := fmt.Sprintf("%x", hash)
	//log.Println("checksum():", path, sum)
	return sum
}

type sig syscall.Signal

func (s sig) String() string {
	return syscall.Signal(s).String()
}

func (s sig) Signal() {}

// MarshalText allows a sig to be represented as a string in JSON objects.
func (s sig) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

var inboundRate, outboundRate uint64

func networkStat() {
	// Total network I/O bytes recieved and sent per second from the system
	// since the start of the system.

	var inbound, outbound, inboundPrev, outboundPrev uint64
	for {
		if tempNet, err := hnet.IOCounters(false); err == nil {
			inbound = tempNet[0].BytesRecv
			outbound = tempNet[0].BytesSent
			inboundRate = inbound - inboundPrev
			outboundRate = outbound - outboundPrev
			inboundPrev = inbound
			outboundPrev = outbound
		}

		time.Sleep(time.Second)
	}
}

// System contains data pertaining to overall system metrics
type System struct {
	CPUPercent float64 `json:"system_cpu_usage"`
	MemPercent float64 `json:"system_mem_usage"`
	Inbound    uint64  `json:"inbound_traffic"`
	Outbound   uint64  `json:"outbound_traffic"`
}

// Event contains data pertaining to the termination of a child process.
type Event struct {
	CheckSum      string      `json:"checksum"`
	UUID          string      `json:"uuid"`
	Time          time.Time   `json:"timestamp"`
	Zone          string      `json:"timezone"`
	IP            string      `json:"public_ip"`
	Status        int         `json:"exit_status"`
	Signal        sig         `json:"signal,omitempty"`
	Trace         interface{} `json:"stack_trace,omitempty"`
	Device        string      `json:"mac_address_hash,omitempty"`
	SystemMetrics System      `json:"system_metrics"`
}

func (e Event) topic() string {
	return envar["EVENT_TOPIC"]
}

func (e *Event) brand() {
	e.UUID = uuid.NewV4().String()
	e.CheckSum = cksum
	e.IP = device.IP
}

func metrics() System {
	var s System

	// System-wide cpu usage since the start of the child process
	if tempCPU, err := cpu.Percent(0, false); err == nil {
		s.CPUPercent = tempCPU[0]
	}

	// System-wide current virtual memory (ram) consumption
	// percentage at the time of child process termination
	if tempMem, err := mem.VirtualMemory(); err == nil {
		s.MemPercent = tempMem.UsedPercent
	}

	s.Inbound = inboundRate
	s.Outbound = outboundRate
	return s
}

func event(evt chan Event, state *os.ProcessState) *Event {
	ws, ok := state.Sys().(syscall.WaitStatus)
	if !ok {
		log.Print("expected type syscall.WaitStatus; non-POSIX system?")
		return nil
	}

	e := Event{
		Device:        device.Mac,
		IP:            device.IP,
		Status:        ws.ExitStatus(),
		SystemMetrics: metrics(),
		Time:          time.Now(),
		Zone:          device.Zone,
	}

	if ws.Signaled() {
		e.Signal = sig(ws.Signal())
	}

	if x, ok := <-evt; ok {
		log.Print("event: got stacktrace")
		e.Signal = x.Signal
		e.Trace = x.Trace
	}
	return &e
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func usage() {
	log.Fatalf("usage: %v command [args ...]\n", os.Args[0])
}

func run(obj chan Object, evt chan Event, cmd *exec.Cmd) (func(), error) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Print("starting child")
	err := cmd.Start()
	if err != nil {
		return func() {}, err
	}

	cpu.Percent(0, false)
	done := make(chan struct{})
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT)

	go func() {
		cmd.Wait()
		obj <- event(evt, cmd.ProcessState)
		done <- struct{}{}
	}()

	return func() {
		defer close(obj)
		for {
			select {
			case s := <-sig:
				log.Print("relaying signal: ", s)
				cmd.Process.Signal(s)
			case <-done:
				log.Print("child exited")
				return
			}
		}
	}, nil
}

// Profile represents arbitrary JSON data from the instrument that can be sent
// to the backend.
type Profile struct {
	CheckSum string      `json:"checksum"`
	IP       string      `json:"public_ip"`
	UUID     string      `json:"uuid"`
	Time     int64       `json:"timestamp"`
	Tree     interface{} `json:"tree"`
}

func (p Profile) topic() string {
	return envar["PROF_TOPIC"]
}

func (p *Profile) brand() {
	p.UUID = uuid.NewV4().String()
	p.CheckSum = cksum
	p.IP = device.IP
	p.Time = time.Now().UnixNano() / 1000000
}

func logs(logger io.Writer) (func(), error) {
	l, err := net.Listen("unixpacket", "log-"+strconv.Itoa(os.Getpid()))
	if err != nil {
		return func() {}, err
	}
	log.Print("logs socket opened")

	done := make(chan error)
	go func() {
		c, err := l.Accept()
		if err != nil {
			done <- err
		}
		log.Print("logs connection accepted")

		t := io.TeeReader(c, logger)
		_, err = ioutil.ReadAll(t)
		done <- err
	}()

	return func() {
		if err := <-done; err != nil {
			log.Print(err)
		}
		log.Print("closing logs socket")
		l.Close()
	}, nil
}

func stacktrace(evt chan Event) (func(), error) {
	s, err := net.Listen("unix", "stacktrace-"+strconv.Itoa(os.Getpid()))
	if err != nil {
		return func() {}, err
	}
	log.Print("stacktrace socket opened")

	done := make(chan error)

	go func() {
		c, err := s.Accept()
		if err != nil {
			done <- err
			return
		}
		log.Print("stacktrace connection accepted")
		line := bufio.NewScanner(c)

		// quits on EOF
		for line.Scan() {
			var s Event
			err := json.Unmarshal(line.Bytes(), &s)
			if err != nil {
				close(evt)
				done <- err
				return
			}
			evt <- s
		}
		close(evt)
		log.Print("stacktrace socket EOF")
		done <- nil
	}()

	return func() {
		// wait for socket relay to finish
		if err := <-done; err != nil {
			log.Print(err)
		}
		log.Print("closing stacktrace socket")
		s.Close()
	}, nil
}

func relay(obj chan Object) (func(), error) {
	s, err := net.Listen("unix", "data-"+strconv.Itoa(os.Getpid()))
	if err != nil {
		return func() {}, err
	}
	log.Print("data socket opened")

	done := make(chan error)

	go func() {
		c, err := s.Accept()
		if err != nil {
			done <- err
		}
		log.Print("data connection accepted")
		line := bufio.NewScanner(c)

		// quits on EOF
		for line.Scan() {
			var p Profile
			err := json.Unmarshal(line.Bytes(), &p.Tree)
			if err != nil {
				done <- err
				return
			}
			obj <- &p
		}
		log.Print("data socket EOF")
		done <- nil
	}()

	return func() {
		// wait for socket relay to finish
		if err := <-done; err != nil {
			log.Print(err)
		}
		log.Print("closing data socket")
		s.Close()
	}, nil
}

func getcerts() map[string][]byte {
	url := envar["BASE_URL"] + "/certificates/"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("apikey", envar["API_KEY"])
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != 200 {
		log.Fatal("wrapper: getcerts: got unexpected status ", resp.Status)
	}

	log.Print("getcerts:", resp.Status)
	// resp.Body implements io.Reader
	// ioutil.ReadAll : io.Reader -> []byte
	// bytes.NewReader : []byte -> bytes.Reader (implements io.ReaderAt)
	// zip.NewReader : io.ReaderAt -> zip.Reader (array of zip.File)
	// zip.Open : zip.File -> io.ReadCloser (implements io.Reader)
	// ioutil.ReadAll : io.Reader -> []byte

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err)
	}
	z, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		log.Panic(err)
	}
	m := make(map[string][]byte)
	for _, f := range z.File {
		rc, err := f.Open()
		if err != nil {
			log.Panic(err)
		}
		cert, err := ioutil.ReadAll(rc)
		if err != nil {
			log.Panic(err)
		}
		m[f.Name] = cert
	}

	filenames := []string{"ck_ca", "ck_cert", "ck_private_key"}
	if len(m) != len(filenames) {
		log.Printf("got zip archive with %v files, expected %v", len(m), len(filenames))
	}

	good := true
	for _, name := range filenames {
		if _, ok := m[name]; !ok {
			log.Printf("could not find cert file named %v", name)
			good = false
		}
	}

	if !good {
		log.Fatal("incorrect certs")
	}
	return m
}

func connect() (sarama.SyncProducer, error) {
	certs := getcerts()

	certpool := x509.NewCertPool()
	certpool.AppendCertsFromPEM(certs["ck_ca"])
	c, err := tls.X509KeyPair(certs["ck_cert"], certs["ck_private_key"])
	check(err)

	tc := tls.Config{
		RootCAs:            certpool,
		ClientAuth:         tls.NoClientCert,
		ClientCAs:          nil,
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{c},
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Net.TLS.Enable = true
	config.Net.TLS.Config = &tc
	config.ClientID = "ProfileTest"

	brokers := strings.Split(envar["BROKERS"], ",")
	return sarama.NewSyncProducer(brokers, config)
}

func produce(obj chan Object, cmd *exec.Cmd) (func(), error) {
	// Create a Kafka producer with the desired config
	p, err := connect()
	if err != nil {
		// bad config or closed client
		return func() {}, err
	}
	log.Println("kafka producer connected")

	done := make(chan error)
	go func() {
		cksum = checksum(cmd.Path)
		if !valid(cksum) {
			//log.Fatal("invalid checksum: ", cksum)
			log.Print("invalid checksum: ", cksum)
		}

		if !device.get() {
			if err := device.post(); err != nil {
				log.Print(err)
			}
		}

		// receive Kafka-bound objects from clients
		for o := range obj {
			o.brand()
			b, err := json.Marshal(o)
			if err != nil {
				done <- err
				return
			}
			log.Printf("producer got %v bytes: %v", len(b), string(b))
			//log.Printf("producer got %v bytes", len(b))
			_, _, err = p.SendMessage(&sarama.ProducerMessage{
				Topic: o.topic(),
				Value: sarama.ByteEncoder(b),
			})
			if err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()

	return func() {
		// wait for kafka producer to finish
		if err := <-done; err != nil {
			log.Print(err)
		}
		log.Print("closing kafka producer")
		p.Close()
	}, nil
}

var cksum string

func valid(sum string) bool {
	ep := envar["BASE_URL"] + "/check_releases/" + sum
	//log.Println("wrapper: release check url:", ep)
	resp, err := http.Get(ep)
	if err != nil {
		log.Panic(err)
	}
	//log.Println("wrapper: valid: response status:", resp.Status)

	switch resp.StatusCode {
	case 200:
		return true
	case 404:
		return false
	default:
		log.Panic("wrapper: valid: got unexpected status ", resp.Status)
	}
	return false
}

// Device contains information about the device that the backend needs to know.
type Device struct {
	Mac   string `json:"mac_address_hash"`
	Zone  string `json:"timezone"`
	AppID string `json:"application"`
	IP    string `json:"-"`
}

func NewDevice() *Device {
	conn, err := net.Dial("udp", "34.235.138.75:80")
	if err != nil {
		log.Fatal(err)
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	zone, _ := time.Now().Zone()
	return &Device{
		Mac:   ifacehash(),
		Zone:  zone,
		AppID: envar["APP_ID"],
		IP:    localAddr.IP.String(),
	}
}

// Determine whether this device is already known by the backend.
func (d *Device) get() bool {
	url := envar["BASE_URL"] + "/devices/?mac_address_hash=" + d.Mac
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("apikey", envar["API_KEY"])
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Device.get() length = ", resp.ContentLength)
	return !(resp.ContentLength <= 2)
}

func ifacehash() string {
	// MAC addresses are generally 6 bytes long
	sum := make([]byte, 6)
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}

	for _, i := range interfaces {
		if bytes.Compare(i.HardwareAddr, nil) == 0 {
			continue
		}
		log.Print(i.HardwareAddr)
		for h, k := range i.HardwareAddr {
			sum[h] += k
		}
	}
	//sum[0]++
	return fmt.Sprintf("%x", string(sum))
}

// Post this device to the backend.
func (d *Device) post() error {
	b, err := json.Marshal(d)
	if err != nil {
		// couldn't marshal json
		log.Fatal(err)
	}
	log.Print(string(b))

	url := envar["BASE_URL"] + "/devices/"
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		// couldn't create this request
		log.Fatal(err)
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("apikey", envar["API_KEY"])

	c := &http.Client{}
	resp, err := c.Do(req)
	log.Print("Device.post() ", resp.Status)
	return err
}

var envar = map[string]string{
	"APP_ID":      "",
	"API_KEY":     "",
	"BASE_URL":    "https://api.auklet.io/v1",
	"BROKERS":     "",
	"PROF_TOPIC":  "",
	"EVENT_TOPIC": "",
	"CA":          "",
	"CERT":        "",
	"PRIVATE_KEY": "",
}

func env() {
	prefix := "AUKLET_"
	ok := true
	for k := range envar {
		v := os.Getenv(prefix + k)
		if v == "" && envar[k] == "" {
			ok = false
			log.Printf("empty envar %v\n", prefix+k)
		} else {
			//log.Print(k, v)
			envar[k] = v
		}
	}
	if !ok {
		log.Fatal("incomplete configuration")
	}
}

var device *Device

func main() {
	logger := os.Stdout
	log.SetOutput(logger)
	log.SetFlags(log.Lmicroseconds)
	log.Printf("Auklet Wrapper version %s (%s)\n", Version, BuildDate)

	env()
	device = NewDevice()
	go networkStat()

	args := os.Args
	if len(args) < 2 {
		usage()
	}
	cmd := exec.Command(args[1], args[2:]...)

	obj := make(chan Object)
	evt := make(chan Event)

	wrelay, err := relay(obj)
	check(err)
	defer wrelay()

	wstack, err := stacktrace(evt)
	check(err)
	defer wstack()

	lc, err := logs(logger)
	check(err)
	defer lc()

	wrun, err := run(obj, evt, cmd)
	check(err)

	wprod, err := produce(obj, cmd)
	check(err)

	wrun()
	wprod()
}
