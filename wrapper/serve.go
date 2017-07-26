package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"

	"github.com/Shopify/sarama"
)

type Node struct {
	Fn, Cs           uint64 `json:",omitempty"`
	Ncalls, Nsamples uint   `json:",omitempty"`
	Callees          []Node `json:",omitempty"`
}

func relay(server net.Listener, producer sarama.AsyncProducer, done chan struct{}) {
	conn, err := server.Accept()
	check(err)

	line := bufio.NewScanner(conn)
	for line.Scan() {
		var n Node
		err := json.Unmarshal(line.Bytes(), &n)
		if err != nil {
			fmt.Println(line.Text())
			panic(err)
		} else {
			fmt.Println("wrapper: got", len(line.Bytes()), "B of valid JSON")
		}
		if producer != nil {
			message := &sarama.ProducerMessage{
				Topic: "sdkTest/sub",
				Value: sarama.ByteEncoder(line.Bytes()),
			}
			producer.Input() <- message
		}
	}

	done <- struct{}{}
}
