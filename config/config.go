// Package config provides configuration parameters for a releaser invocation.
package config

import (
	"log"
	"os"
)

// A Config represents parameters of a releaser invocation.
type Config struct {
	BaseURL string
	APIKey  string
	AppID   string
}

// ReleaseBuild creates a Config as would be required in a production
// environment. The base URL is hardcoded in this configuration and cannot be
// overridden by the end user.
func ReleaseBuild() Config {
	return Config{
		BaseURL: StaticBaseURL,
		APIKey:  os.Getenv("AUKLET_API_KEY"),
		AppID:   os.Getenv("AUKLET_APP_ID"),
	}
}

// LocalBuild returns a configuration defined solely from the environment.
func LocalBuild() (c Config) {
	c = Config{
		BaseURL: os.Getenv("AUKLET_BASE_URL"),
		APIKey:  os.Getenv("AUKLET_API_KEY"),
		AppID:   os.Getenv("AUKLET_APP_ID"),
	}
	return
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
