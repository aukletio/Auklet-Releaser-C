package config

import (
	"log"
	"os"
)

type Config struct {
	BaseUrl string
	APIKey string
	AppId string
}

func Production() Config {
	return Config{
		BaseUrl: "https://api.auklet.io/private",
		APIKey: os.Getenv("AUKLET_API_KEY"),
		AppId: os.Getenv("AUKLET_API_KEY"),
	}
}

func FromEnv() (c Config) {
	c = Config{
		BaseUrl: os.Getenv("AUKLET_BASE_URL"),
		APIKey: os.Getenv("AUKLET_API_KEY"),
		AppId: os.Getenv("AUKLET_API_KEY"),
	}
	return
}

func (c Config) Valid() (ok bool) {
	ok = true
	if "" == c.BaseUrl {
		log.Printf("warning: empty BASE_URL")
		ok = false
	}
	if "" == c.APIKey {
		log.Printf("warning: empty API_KEY")
		ok = false
	}
	if "" == c.AppId {
		log.Printf("warning: empty APP_ID")
		ok = false
	}
	return
}
