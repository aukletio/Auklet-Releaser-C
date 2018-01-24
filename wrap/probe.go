package main

import (
	"encoding/json"
	"log"
)

func probe(prefix string, v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		log.Print(err)
	}
	log.Println(prefix, string(b))
}
