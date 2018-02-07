package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/sha512"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	stdlog "log"
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
	"github.com/rdegges/go-ipify"
	"github.com/satori/go.uuid"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	hnet "github.com/shirou/gopsutil/net"
)

// BuildDate is provided at compile-time; DO NOT MODIFY.
var BuildDate = "no timestamp"

// Version is provided at compile-time; DO NOT MODIFY.
var Version = "local-build"

type logLevel string

type logWriter struct {
	level logLevel
	send  SendFn
}

func (lw *logWriter) Write(p []byte) (n int, err error) {
	if envar["DUMP"] == "true" {
		fmt.Print(string(p))
	}
	return len(p), lw.send(&log{
		Level:   lw.level,
		Message: string(p),
	})
}

// Object represents something that can be sent to the backend. It must have a
// topic and implement a brand() method that fills UUID and checksum fields.
type Object interface {
	topic() string
	brand(string)
}

func checksum(path string) (sum string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	h := sha512.New512_224()
	if _, err = io.Copy(h, f); err != nil {
		return
	}
	sum = fmt.Sprintf("%x", h.Sum(nil))
	return
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

func networkStat() { // inboundRate outBoundRate
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

type Common struct {
	CheckSum string `json:"checksum"`
	IP       string `json:"public_ip"`
	UUID     string `json:"uuid"`
}

// Event contains data pertaining to the termination of a child process.
type Event struct {
	Common
	Time          time.Time   `json:"timestamp"`
	Zone          string      `json:"timezone"`
	Status        int         `json:"exit_status"`           // waitstatus
	Signal        sig         `json:"signal,omitempty"`      // waitstatus | json
	Trace         interface{} `json:"stack_trace,omitempty"` // json
	Device        string      `json:"mac_address_hash,omitempty"`
	SystemMetrics System      `json:"system_metrics"`
}

func (e Event) topic() string {
	return envar["EVENT_TOPIC"]
}

func (e *Event) brand(cksum string) {
	e.UUID = uuid.NewV4().String()
	e.CheckSum = cksum
	e.IP = device.IP

	e.Device = device.Mac

	e.SystemMetrics = metrics()
	e.Time = time.Now()
	e.Zone = device.Zone
}

func metrics() System { // inboundRate outboundRate
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

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func usage() {
	stdlog.Fatalf("usage: %v command [args ...]\n", os.Args[0])
}

func relaysigs(cmd *exec.Cmd) {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT)
	for s := range sig {
		debug.Print("relaying signal: ", s)
		cmd.Process.Signal(s)
	}
}

type SendFn func(Object) error

// Profile represents a profile tree to be sent to Kafka.
type Profile struct {
	Common
	Time  int64           `json:"timestamp"`
	Tree  json.RawMessage `json:"tree"`
	AppID string          `json:"app_id"`
}

func (p Profile) topic() string {
	return envar["PROF_TOPIC"]
}

func (p *Profile) brand(cksum string) {
	p.UUID = uuid.NewV4().String()
	p.CheckSum = cksum
	p.IP = device.IP
	p.AppID = envar["APP_ID"]
	p.Time = time.Now().UnixNano() / 1000000
}

type log struct {
	Level   logLevel `json:"level"`
	Message string   `json:"message"`
}

func (l *log) topic() string {
	return envar["LOG_TOPIC"]
}

func (l *log) brand(_ string) {}

func objectify(b []byte, wait WaitFn, send SendFn) (done bool, err error) {
	j := struct {
		Type string
		Data json.RawMessage
	}{}
	err = json.Unmarshal(b, &j)
	if err != nil {
		return
	}
	var o Object
	switch j.Type {
	case "log":
		o = &log{}
	case "event":
		ws := wait()
		done = true
		o = &Event{Status: ws.ExitStatus()}
	case "profile":
		o = &Profile{}
	default:
		err = errors.New(fmt.Sprintf("objectify: couldn't match %v\n", j.Type))
		return
	}
	err = json.Unmarshal(j.Data, o)
	if err != nil {
		return
	}
	send(o)
	return
}

type WaitFn func() syscall.WaitStatus

func relay(s net.Listener, send SendFn, cmd *exec.Cmd) (err error) {
	err = cmd.Start()
	if err != nil {
		return
	}
	info.Print("child started")
	wait := func() syscall.WaitStatus {
		cmd.Wait()
		info.Print("child exited")
		return cmd.ProcessState.Sys().(syscall.WaitStatus)
	}
	go relaysigs(cmd)
	cpu.Percent(0, false)
	c, err := s.Accept()
	if err != nil {
		return
	}
	info.Printf("socket connection accepted")
	line := bufio.NewScanner(c)
	for line.Scan() {
		done, err := objectify(line.Bytes(), wait, send)
		if err != nil {
			return err
		}
		if done {
			// The instrument sent a stacktrace, so we don't need to
			// wait for EOF; return immediately.
			return nil
		}
	}
	info.Printf("socket EOF")
	ws := wait()
	e := &Event{
		Status: ws.ExitStatus(),
	}
	if ws.Signaled() {
		e.Signal = sig(ws.Signal())
	}
	err = send(e)
	return
}

var info, debug, fatal *stdlog.Logger

func loginit(send SendFn) {
	for _, l := range []struct {
		lg    **stdlog.Logger
		level logLevel
	}{
		{&info, "info"},
		{&debug, "debug"},
		{&fatal, "fatal"},
	} {
		*l.lg = stdlog.New(&logWriter{
			level: l.level,
			send:  send,
		}, "", stdlog.Lmicroseconds)
	}
}

func manage(cmd *exec.Cmd) (obj chan Object) {
	obj = make(chan Object, 10)
	send := func(o Object) (err error) {
		t := time.NewTimer(20 * time.Second)
		select {
		case obj <- o:
			t.Stop()
		case <-t.C:
			err = errors.New("obj <- o timed out")
		}
		return
	}
	loginit(send)
	addr := "/tmp/auklet-" + strconv.Itoa(os.Getpid())
	s, err := net.Listen("unixpacket", addr)
	check(err)
	info.Printf("%v opened", addr)
	go func() {
		var err error
		defer func() {
			if err != nil {
				info.Println(err)
			}
			info.Printf("%v closing", addr)
			s.Close()
			close(obj)
		}()
		err = relay(s, send, cmd)
	}()
	return
}

func getcerts() (m map[string][]byte, err error) {
	url := envar["BASE_URL"] + "/certificates/"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Add("apikey", envar["API_KEY"])
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		format := "getcerts: got unexpected status %v"
		err = errors.New(fmt.Sprintf(format, resp.Status))
		return
	}

	// resp.Body implements io.Reader
	// ioutil.ReadAll : io.Reader -> []byte
	// bytes.NewReader : []byte -> bytes.Reader (implements io.ReaderAt)
	// zip.NewReader : io.ReaderAt -> zip.Reader (array of zip.File)
	// zip.Open : zip.File -> io.ReadCloser (implements io.Reader)
	// ioutil.ReadAll : io.Reader -> []byte

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	z, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return
	}
	m = make(map[string][]byte)
	for _, f := range z.File {
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		cert, err := ioutil.ReadAll(rc)
		if err != nil {
			return nil, err
		}
		m[f.Name] = cert
	}

	filenames := []string{"ck_ca", "ck_cert", "ck_private_key"}
	if len(m) != len(filenames) {
		format := "got zip archive with %v files, expected %v"
		err = errors.New(fmt.Sprintf(format, len(m), len(filenames)))
		return nil, err
	}

	good := true
	for _, name := range filenames {
		if _, ok := m[name]; !ok {
			fatal.Printf("could not find cert file named %v", name)
			good = false
		}
	}

	if !good {
		err = errors.New("incorrect certs")
		return nil, err
	}
	return
}

func connect() (p sarama.SyncProducer, err error) {
	certs, err := getcerts()
	if err != nil {
		return
	}

	certpool := x509.NewCertPool()
	certpool.AppendCertsFromPEM(certs["ck_ca"])
	c, err := tls.X509KeyPair(certs["ck_cert"], certs["ck_private_key"])
	if err != nil {
		return
	}

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

type Producer struct {
	CheckSum string
	Dev      *Device
	P        sarama.SyncProducer
}

func NewProducer(path string) (p *Producer, err error) {
	cksum, err := checksum(path)
	if err != nil {
		return
	}
	ok, err := valid(cksum)
	if err != nil {
		return
	}
	if !ok {
		err = errors.New(fmt.Sprintf("checksum %v... not released", cksum[:10]))
		return
	}
	ok, err = device.get()
	if err != nil {
		return
	}
	if !ok {
		err = device.post()
		if err != nil {
			return
		}
	}
	sp, err := connect()
	if err != nil {
		return // bad config or closed client
	}
	p = &Producer{
		P:        sp,
		CheckSum: cksum,
		Dev:      device,
	}
	return
}

func (p *Producer) Close() {
	p.P.Close()
}

func (p *Producer) produce(obj <-chan Object) (err error) {
	defer func() {
		if err != nil {
			stdlog.Print(err)
		}
		p.Close()
	}()
	info.Println("kafka producer connected")
	// receive Kafka-bound objects from clients
	for o := range obj {
		o.brand(p.CheckSum)
		b, err := json.Marshal(o)
		if err != nil {
			return err
		}
		if envar["DUMP"] == "true" {
			fmt.Printf("producer got %v bytes: %v\n", len(b), string(b))
		}
		_, _, err = p.P.SendMessage(&sarama.ProducerMessage{
			Topic: o.topic(),
			Value: sarama.ByteEncoder(b),
		})
		if err != nil {
			return err
		}
	}
	return
}

func valid(sum string) (ok bool, err error) {
	ep := envar["BASE_URL"] + "/check_releases/" + sum
	//stdlog.Println("wrapper: release check url:", ep)
	resp, err := http.Get(ep)
	if err != nil {
		return
	}
	//stdlog.Println("wrapper: valid: response status:", resp.Status)

	switch resp.StatusCode {
	case 200:
		// released
		ok = true
	case 404:
		// not released
		ok = false
	// 500 happens if the backend is broken teehee
	default:
		format := "valid: got unexpected status %v"
		err = errors.New(fmt.Sprintf(format, resp.Status))
	}
	return
}

// Device contains information about the device that the backend needs to know.
type Device struct {
	Mac   string `json:"mac_address_hash"`
	Zone  string `json:"timezone"`
	AppID string `json:"application"`
	IP    string `json:"-"`
}

func NewDevice() *Device {
	zone, _ := time.Now().Zone()
	d := &Device{
		Mac:   ifacehash(),
		Zone:  zone,
		AppID: envar["APP_ID"],
		IP:    getip(),
	}
	go func() { // d
		for _ = range time.Tick(5 * time.Minute) {
			d.IP = getip()
		}
	}()
	return d
}

func getip() string {
	ip, err := ipify.GetIp()
	if err != nil {
		debug.Print(err)
	}
	return ip
}

// Determine whether this device is already known by the backend.
func (d *Device) get() (ok bool, err error) {
	url := envar["BASE_URL"] + "/devices/?mac_address_hash=" + d.Mac
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Add("apikey", envar["API_KEY"])
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return
	}

	debug.Print("Device.get() length = ", resp.ContentLength)
	return !(resp.ContentLength <= 2), nil
}

func ifacehash() string {
	// MAC addresses are generally 6 bytes long
	sum := make([]byte, 6)
	interfaces, err := net.Interfaces()
	if err != nil {
		stdlog.Fatal(err)
	}

	for _, i := range interfaces {
		if bytes.Compare(i.HardwareAddr, nil) == 0 {
			continue
		}
		//stdlog.Print(i.HardwareAddr)
		for h, k := range i.HardwareAddr {
			sum[h] += k
		}
	}
	//sum[0]++
	return fmt.Sprintf("%x", string(sum))
}

// Post this device to the backend.
func (d *Device) post() (err error) {
	b, err := json.Marshal(d)
	if err != nil {
		return
	}
	debug.Print(string(b))

	url := envar["BASE_URL"] + "/devices/"
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("apikey", envar["API_KEY"])

	c := &http.Client{}
	resp, err := c.Do(req)
	debug.Print("Device.post() ", resp.Status)
	return
}

var envar = map[string]string{
	"DUMP":        "false",
	"APP_ID":      "",
	"API_KEY":     "",
	"BASE_URL":    "https://api.auklet.io/v1",
	"BROKERS":     "",
	"PROF_TOPIC":  "",
	"EVENT_TOPIC": "",
	"LOG_TOPIC":   "",
}

func env() {
	prefix := "AUKLET_"
	ok := true
	for k := range envar {
		v := os.Getenv(prefix + k)
		if v == "" && envar[k] == "" {
			ok = false
			stdlog.Printf("empty envar %v\n", prefix+k)
		} else {
			//stdlog.Print(k, v)
			envar[k] = v
		}
	}
	if !ok {
		stdlog.Fatal("incomplete configuration")
	}
}

var device *Device

func main() {
	logger := os.Stdout
	stdlog.SetOutput(logger)
	stdlog.SetFlags(stdlog.Lmicroseconds)
	stdlog.Printf("Auklet Wrapper version %s (%s)\n", Version, BuildDate)

	env()
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
	}
	device = NewDevice()
	go networkStat()

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	obj := manage(cmd)
	p, err := NewProducer(cmd.Path)
	check(err)
	err = p.produce(obj)
	check(err)
}
