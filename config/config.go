// Package config provides configuration parameters for a releaser invocation.
package config

import (
	"log"
	"os"
)

// Production defines the base URL for the production environment.
const Production = "https://api.auklet.io"

// A Config represents parameters of a releaser invocation.
type Config struct {
	BaseURL string
	APIKey  string
	AppID   string
}

// GetConfig returns a config object whose BaseURL is dependent upon CLI args
// or env vars.
func GetConfig(fromcli string) Config {
	var baseURL string
	if fromcli != "" {
		baseURL = fromcli
	} else {
		fromenv := os.Getenv("AUKLET_BASE_URL")
		if fromenv != "" {
			baseURL = fromenv
		} else {
			baseURL = Production
		}
	}
	return Config{
		BaseURL: baseURL,
		APIKey:  os.Getenv("AUKLET_API_KEY"),
		AppID:   os.Getenv("AUKLET_APP_ID"),
	}
}

// Valid returns true if c has no empty fields, false otherwise.
func (c Config) Valid() (ok bool) {
	ok = true
	if "" == c.BaseURL {
		log.Printf("warning: empty BASE_URL")
		ok = false
	}
	if "" == c.APIKey {
		log.Printf("warning: empty API_KEY")
		ok = false
	}
	if "" == c.AppID {
		log.Printf("warning: empty APP_ID")
		ok = false
	}
	return
}
