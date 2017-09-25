package main

import (
	"crypto/sha512"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func checksum(path string) string {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := sha512.New512_224()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	hash := h.Sum(nil)
	return fmt.Sprintf("%x", hash)
}

func valid(sum string) bool {
	endpoint := "https://api-staging.auklet.io/v1/check_releases/"

	resp, err := http.Get(endpoint + sum)
	fmt.Println("wrapper: valid: response status:", resp.Status)
	if err != nil {
		log.Fatal(err)
	}

	switch resp.StatusCode {
	case 200:
		return true
	case 404:
		return false
	default:
		log.Fatal("wrapper: valid: got unexpected status ", resp.Status)
	}
	return false
}
