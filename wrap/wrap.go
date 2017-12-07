package main

import (
	"bufio"
	"bytes"
	"crypto/sha512"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
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

type frame struct {
	Fn uint64 `json:"fn,omitempty"`
	Cs uint64 `json:"cs,omitempty"`
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

// DeviceIP gets the public IP address of the device
func DeviceIP() string {
	conn, err := net.Dial("udp", "34.235.138.75:80")
	if err != nil {
		log.Println("could not get the device IP")
		return ""
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
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
	CheckSum      string    `json:"checksum"`
	UUID          string    `json:"uuid"`
	Time          time.Time `json:"timestamp"`
	Zone          string    `json:"timezone"`
	IP            string    `json:"public_ip"`
	Status        int       `json:"exit_status"`
	Signal        sig       `json:"signal,omitempty"`
	Trace         []frame   `json:"stack_trace,omitempty"`
	Device        string    `json:"mac_address_hash,omitempty"`
	SystemMetrics System    `json:"system_metrics"`
}

func (e Event) topic() string {
	return envar["EVENT_TOPIC"]
}

func (e *Event) brand() {
	e.UUID = uuid.NewV4().String()
	e.CheckSum = cksum
	e.IP = deviceIP
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

	local := time.Now()
	zone, _ := local.Zone()
	e := Event{
		Device:        hash,
		IP:            deviceIP,
		Status:        ws.ExitStatus(),
		SystemMetrics: metrics(),
		Time:          local,
		Zone:          zone,
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

func run(obj chan Object, evt chan Event, cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Print("starting child")
	err := cmd.Start()
	if err != nil {
		panic(err)
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
}

// Node is used by json.Unmarshal() to check JSON generated by the instrument.
type Node struct {
	CheckSum string `json:"checksum,omitempty"`
	IP       string `json:"public_ip,omitempty"`
	UUID     string `json:"uuid,omitempty"`
	frame
	Ncalls   uint   `json:"ncalls,omitempty"`
	Nsamples uint   `json:"nsamples,omitempty"`
	Callees  []Node `json:"callees,omitempty"`
}

func (n Node) topic() string {
	return envar["PROF_TOPIC"]
}

func (n *Node) brand() {
	n.UUID = uuid.NewV4().String()
	n.CheckSum = cksum
	n.IP = deviceIP
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
			var n Node
			err := json.Unmarshal(line.Bytes(), &n)
			if err != nil {
				done <- err
				return
			}
			obj <- &n
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

func decode(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s)
	check(err)
	return b
}

func connect() (sarama.SyncProducer, error) {
	ca := decode(envar["CA"])
	cert := decode(envar["CERT"])
	key := decode(envar["PRIVATE_KEY"])

	certpool := x509.NewCertPool()
	certpool.AppendCertsFromPEM(ca)
	c, err := tls.X509KeyPair(cert, key)
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

func produce(obj chan Object) (func(), error) {
	// Create a Kafka producer with the desired config
	p, err := connect()
	if err != nil {
		// bad config or closed client
		return func() {}, err
	}
	log.Println("kafka producer connected")

	done := make(chan error)
	go func() {
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

// Device contains information that need to be posted to device endpoint
type Device struct {
	Mac   string `json:"mac_address_hash,omitempty"`
	Zone  string `json:"timezone,omitempty"`
	AppID string `json:"application,omitempty"`
}

var hash string

func postDevice() error {
	//Mac addresses are generally 6 bytes long
	sum := make([]byte, 6)
	var url string
	apikey := envar["API_KEY"]
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}

	for _, i := range interfaces {
		if bytes.Compare(i.HardwareAddr, nil) == 0 {
			continue
		}
		for h, k := range i.HardwareAddr {
			sum[h] += k
		}
	}
	hash = fmt.Sprintf("%x", string(sum))

	zone, _ := time.Now().Zone()
	d := Device{
		Mac:   hash,
		Zone:  zone,
		AppID: envar["APP_ID"],
	}

	b, err := json.Marshal(d)
	url = envar["BASE_URL"] + "/devices/" + hash

	res, err := http.Get(url)

	if res.StatusCode != 200 {
		url = envar["BASE_URL"] + "/devices"
		client := &http.Client{
			Timeout: 5 * time.Second,
		}

		req, err := http.NewRequest("POST", url, bytes.NewReader(b))
		if err != nil {
			return err
		}

		req.Header.Add("content-type", "application/json")
		req.Header.Add("apikey", apikey)

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		log.Print("postDevice:", resp.Status)
	}
	// If we get to this point, whatever the response code is we do not return any error
	return nil
}

var envar = map[string]string{
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
	for k, _ := range envar {
		v := os.Getenv(prefix + k)
		if v == "" && envar[k] == "" {
			ok = false
			log.Printf("empty envar %v\n", prefix+k)
		} else {
			envar[k] = v
		}
	}
	if !ok {
		log.Fatal("incomplete configuration")
	}
}

var deviceIP string

func main() {
	logger := os.Stdout
	log.SetOutput(logger)
	log.Printf("Auklet Wrapper version %s (%s)\n", Version, BuildDate)

	env()
	deviceIP = DeviceIP()
	go networkStat()

	args := os.Args
	if len(args) < 2 {
		usage()
	}
	cmd := exec.Command(args[1], args[2:]...)

	cksum = checksum(cmd.Path)
	if !valid(cksum) {
		//log.Fatal("invalid checksum: ", cksum)
		log.Print("invalid checksum: ", cksum)
	}

	devicePosted := postDevice()
	if devicePosted != nil {
		log.Print("Failed to post device object")
	}

	obj := make(chan Object)
	evt := make(chan Event)

	wprod, err := produce(obj)
	check(err)
	defer wprod()

	wrelay, err := relay(obj)
	check(err)
	defer wrelay()

	wstack, err := stacktrace(evt)
	check(err)
	defer wstack()

	lc, err := logs(logger)
	check(err)
	defer lc()

	run(obj, evt, cmd)
	close(obj)
}
