package config

import (
	"testing"
)

func TestValid(t *testing.T) {
	c := Config{}
	if c.Valid() {
		t.Fail()
	}
	c.BaseUrl = "not empty"
	if c.Valid() {
		t.Fail()
	}
	c.AppId = "not empty"
	if c.Valid() {
		t.Fail()
	}
	c.APIKey = "not empty"
	if !c.Valid() {
		t.Fail()
	}
}
