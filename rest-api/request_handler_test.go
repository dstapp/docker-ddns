package main

import (
    "testing"
    "net/http"
)

func TestBuildWebserviceResponseFromRequestToReturnValidObject(t *testing.T) {
    var appConfig = &Config{}
    appConfig.SharedSecret = "changeme"

    req, _ := http.NewRequest("POST", "/update?secret=changeme&domain=foo&addr=1.2.3.4", nil)
    result := BuildWebserviceResponseFromRequest(req, appConfig)

    if result.Success != true {
        t.Fatalf("Expected WebserviceResponse.Success to be true")
    }

    if result.Domain != "foo" {
        t.Fatalf("Expected WebserviceResponse.Domain to be foo")
    }

    if result.Address != "1.2.3.4" {
        t.Fatalf("Expected WebserviceResponse.Address to be 1.2.3.4")
    }

    if result.AddrType != "A" {
        t.Fatalf("Expected WebserviceResponse.AddrType to be A")
    }
}

func TestBuildWebserviceResponseFromRequestToReturnInvalidObjectWhenNoSecretIsGiven(t *testing.T) {
    var appConfig = &Config{}
    appConfig.SharedSecret = "changeme"

    req, _ := http.NewRequest("POST", "/update", nil)
    result := BuildWebserviceResponseFromRequest(req, appConfig)

    if result.Success != false {
        t.Fatalf("Expected WebserviceResponse.Success to be false")
    }
}

func TestBuildWebserviceResponseFromRequestToReturnInvalidObjectWhenInvalidSecretIsGiven(t *testing.T) {
    var appConfig = &Config{}
    appConfig.SharedSecret = "changeme"

    req, _ := http.NewRequest("POST", "/update?secret=foo", nil)
    result := BuildWebserviceResponseFromRequest(req, appConfig)

    if result.Success != false {
        t.Fatalf("Expected WebserviceResponse.Success to be false")
    }
}

func TestBuildWebserviceResponseFromRequestToReturnInvalidObjectWhenNoDomainIsGiven(t *testing.T) {
    var appConfig = &Config{}
    appConfig.SharedSecret = "changeme"

    req, _ := http.NewRequest("POST", "/update?secret=changeme", nil)
    result := BuildWebserviceResponseFromRequest(req, appConfig)

    if result.Success != false {
        t.Fatalf("Expected WebserviceResponse.Success to be false")
    }
}

func TestBuildWebserviceResponseFromRequestToReturnInvalidObjectWhenNoAddressIsGiven(t *testing.T) {
    var appConfig = &Config{}
    appConfig.SharedSecret = "changeme"

    req, _ := http.NewRequest("POST", "/update?secret=changeme&domain=foo", nil)
    result := BuildWebserviceResponseFromRequest(req, appConfig)

    if result.Success != false {
        t.Fatalf("Expected WebserviceResponse.Success to be false")
    }
}

func TestBuildWebserviceResponseFromRequestToReturnInvalidObjectWhenInvalidAddressIsGiven(t *testing.T) {
    var appConfig = &Config{}
    appConfig.SharedSecret = "changeme"

    req, _ := http.NewRequest("POST", "/update?secret=changeme&domain=foo&addr=1.41:2", nil)
    result := BuildWebserviceResponseFromRequest(req, appConfig)

    if result.Success != false {
        t.Fatalf("Expected WebserviceResponse.Success to be false")
    }
}
