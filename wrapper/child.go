package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/Shopify/sarama"
)

// Run a command and report when it exits.
func run(cmd *exec.Cmd, msg chan sarama.ProducerMessage, wg *sync.WaitGroup) {
	defer wg.Done()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	check(err)
	err = cmd.Wait()
	fmt.Println(err)

	b, e := json.Marshal(err)
	check(e)
	log.Printf(string(b))

	staging := "f9l0-events"
	msg <- sarama.ProducerMessage{
		Topic: staging,
		Value: sarama.ByteEncoder(b),
	}
	close(msg)
}
