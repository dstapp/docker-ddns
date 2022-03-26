package main

import (
	"github.com/dprandzioch/docker-ddns/rest-api/ipparser"
	"testing"
)

func TestValidIP4ToReturnTrueOnValidAddress(t *testing.T) {
	result := ipparser.ValidIP4("1.2.3.4")

	if result != true {
		t.Fatalf("Expected ValidIP(1.2.3.4) to be true but got false")
	}
}

func TestValidIP4ToReturnFalseOnInvalidAddress(t *testing.T) {
	result := ipparser.ValidIP4("abcd")

	if result == true {
		t.Fatalf("Expected ValidIP(abcd) to be false but got true")
	}
}

func TestValidIP4ToReturnFalseOnEmptyAddress(t *testing.T) {
	result := ipparser.ValidIP4("")

	if result == true {
		t.Fatalf("Expected ValidIP() to be false but got true")
	}
}
