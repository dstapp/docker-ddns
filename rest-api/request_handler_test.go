package main

import (
	"encoding/base64"
	"net/http"
	"testing"
)

var defaultExtractor = defaultRequestDataExtractor{}
var dynExtractor = dynRequestDataExtractor{}

func TestBuildWebserviceResponseFromRequestToReturnValidObject(t *testing.T) {
	var appConfig = &Config{}

	req, _ := http.NewRequest("GET", "/update?secret=changeme&domain=foo&addr=1.2.3.4", nil)
	result := BuildWebserviceResponseFromRequest(req, appConfig, defaultExtractor)

	if result.Success != true {
		t.Fatalf("Expected WebserviceResponse.Success to be true")
	}

	if result.Domain != "foo" {
		t.Fatalf("Expected WebserviceResponse.Domain to be foo")
	}

	if result.Addresses[0].Address != "1.2.3.4" {
		t.Fatalf("Expected WebserviceResponse.Address to be 1.2.3.4")
	}

	if result.Addresses[0].AddrType != "A" {
		t.Fatalf("Expected WebserviceResponse.AddrType to be A")
	}
}

func TestBuildWebserviceResponseFromRequestWithXRealIPHeaderToReturnValidObject(t *testing.T) {
	var appConfig = &Config{}

	req, _ := http.NewRequest("GET", "/update?secret=changeme&domain=foo", nil)
	req.Header.Add("X-Real-Ip", "1.2.3.4")
	result := BuildWebserviceResponseFromRequest(req, appConfig, defaultExtractor)

	if result.Success != true {
		t.Fatalf("Expected WebserviceResponse.Success to be true")
	}

	if result.Domain != "foo" {
		t.Fatalf("Expected WebserviceResponse.Domain to be foo")
	}

	if result.Addresses[0].Address != "1.2.3.4" {
		t.Fatalf("Expected WebserviceResponse.Address to be 1.2.3.4")
	}

	if result.Addresses[0].AddrType != "A" {
		t.Fatalf("Expected WebserviceResponse.AddrType to be A")
	}
}

func TestBuildWebserviceResponseFromRequestWithXForwardedForHeaderToReturnValidObject(t *testing.T) {
	var appConfig = &Config{}

	req, _ := http.NewRequest("GET", "/update?secret=changeme&domain=foo", nil)
	req.Header.Add("X-Forwarded-For", "1.2.3.4")
	result := BuildWebserviceResponseFromRequest(req, appConfig, defaultExtractor)

	if result.Success != true {
		t.Fatalf("Expected WebserviceResponse.Success to be true")
	}

	if result.Domain != "foo" {
		t.Fatalf("Expected WebserviceResponse.Domain to be foo but was %s", result.Domain)
	}

	if result.Addresses[0].Address != "1.2.3.4" {
		t.Fatalf("Expected WebserviceResponse.Address to be 1.2.3.4 but was %s", result.Addresses[0].Address)
	}

	if result.Addresses[0].AddrType != "A" {
		t.Fatalf("Expected WebserviceResponse.AddrType to be A but was %s", result.Addresses[0].AddrType)
	}
}

func TestBuildWebserviceResponseFromRequestToReturnInvalidObjectWhenNoSecretIsGiven(t *testing.T) {
	var appConfig = &Config{}

	req, _ := http.NewRequest("GET", "/update", nil)
	result := BuildWebserviceResponseFromRequest(req, appConfig, defaultExtractor)

	if result.Success != false {
		t.Fatalf("Expected WebserviceResponse.Success to be false")
	}
}

func TestBuildWebserviceResponseFromRequestToReturnInvalidObjectWhenInvalidSecretIsGiven(t *testing.T) {
	var appConfig = &Config{}

	req, _ := http.NewRequest("GET", "/update?secret=foo", nil)
	result := BuildWebserviceResponseFromRequest(req, appConfig, defaultExtractor)

	if result.Success != false {
		t.Fatalf("Expected WebserviceResponse.Success to be false")
	}
}

func TestBuildWebserviceResponseFromRequestToReturnInvalidObjectWhenNoDomainIsGiven(t *testing.T) {
	var appConfig = &Config{}

	req, _ := http.NewRequest("GET", "/update?secret=changeme", nil)
	result := BuildWebserviceResponseFromRequest(req, appConfig, defaultExtractor)

	if result.Success != false {
		t.Fatalf("Expected WebserviceResponse.Success to be false")
	}
}

func TestBuildWebserviceResponseFromRequestWithMultipleDomains(t *testing.T) {
	var appConfig = &Config{}

	req, _ := http.NewRequest("GET", "/update?secret=changeme&domain=foo,bar&addr=1.2.3.4", nil)
	result := BuildWebserviceResponseFromRequest(req, appConfig, defaultExtractor)

	if result.Success != true {
		t.Fatalf("Expected WebserviceResponse.Success to be true")
	}

	if len(result.Domains) != 2 {
		t.Fatalf("Expected WebserviceResponse.Domains length to be 2")
	}

	if result.Domains[0] != "foo" {
		t.Fatalf("Expected WebserviceResponse.Domains[0] to equal 'foo'")
	}

	if result.Domains[1] != "bar" {
		t.Fatalf("Expected WebserviceResponse.Domains[1] to equal 'bar'")
	}
}

func TestBuildWebserviceResponseFromRequestWithMalformedMultipleDomains(t *testing.T) {
	var appConfig = &Config{}

	req, _ := http.NewRequest("GET", "/update?secret=changeme&domain=foo,&addr=1.2.3.4", nil)
	result := BuildWebserviceResponseFromRequest(req, appConfig, defaultExtractor)

	if result.Success != false {
		t.Fatalf("Expected WebserviceResponse.Success to be false")
	}
}

func TestBuildWebserviceResponseFromRequestToReturnInvalidObjectWhenNoAddressIsGiven(t *testing.T) {
	var appConfig = &Config{}

	req, _ := http.NewRequest("POST", "/update?secret=changeme&domain=foo", nil)
	result := BuildWebserviceResponseFromRequest(req, appConfig, defaultExtractor)

	if result.Success != false {
		t.Fatalf("Expected WebserviceResponse.Success to be false")
	}
}

func TestBuildWebserviceResponseFromRequestToReturnInvalidObjectWhenInvalidAddressIsGiven(t *testing.T) {
	var appConfig = &Config{}

	req, _ := http.NewRequest("GET", "/update?secret=changeme&domain=foo&addr=1.41:2", nil)
	result := BuildWebserviceResponseFromRequest(req, appConfig, defaultExtractor)

	if result.Success != false {
		t.Fatalf("Expected WebserviceResponse.Success to be false")
	}
}

func TestBuildWebserviceResponseFromRequestToReturnValidObjectWithDynExtractor(t *testing.T) {
	var appConfig = &Config{}

	req, _ := http.NewRequest("GET", "/nic/update?hostname=foo&myip=1.2.3.4", nil)
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("username:changeme")))

	result := BuildWebserviceResponseFromRequest(req, appConfig, dynExtractor)

	if result.Success != true {
		t.Fatalf("Expected WebserviceResponse.Success to be true")
	}

	if result.Domain != "foo" {
		t.Fatalf("Expected WebserviceResponse.Domain to be foo but was %s", result.Domain)
	}

	if result.Addresses[0].Address != "1.2.3.4" {
		t.Fatalf("Expected WebserviceResponse.Address to be 1.2.3.4 but was %s", result.Addresses[0].Address)
	}

	if result.Addresses[0].AddrType != "A" {
		t.Fatalf("Expected WebserviceResponse.AddrType to be A but was %s", result.Addresses[0].AddrType)
	}
}

func TestBuildWebserviceResponseFromRequestToReturnInvalidObjectWhenNoSecretIsGivenWithDynExtractor(t *testing.T) {
	var appConfig = &Config{}

	req, _ := http.NewRequest("GET", "/nic/update", nil)
	result := BuildWebserviceResponseFromRequest(req, appConfig, dynExtractor)

	if result.Success != false {
		t.Fatalf("Expected WebserviceResponse.Success to be false")
	}
}
