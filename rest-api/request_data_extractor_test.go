package main

import (
	"net/http"
	"testing"
)

func verify(t *testing.T, r *http.Request, extractor requestDataExtractor, domain string, expected RecordUpdateRequest) {
	rru := RecordUpdateRequest{
		ddnsKeyName: extractor.DdnsKeyName(r, domain),
		zone:        extractor.Zone(r, domain),
		fqdn:        extractor.Fqdn(r, domain),
	}
	if rru.zone != expected.zone {
		t.Fatalf("Zone not configured but not empty: %s != %s", rru.zone, expected.zone)
	}

	if rru.fqdn != expected.fqdn {
		t.Fatalf("Wrong fqdn: %s != %s", rru.fqdn, expected.fqdn)
	}

	if rru.ddnsKeyName != expected.ddnsKeyName {
		t.Fatalf("Wrong ddnskeyname: %s != %s", rru.ddnsKeyName, expected.ddnsKeyName)
	}
}

func TestExtractorUnconfiguredZone(t *testing.T) {
	var e = defaultRequestDataExtractor{appConfig: &Config{
		Zone: "",
	}}

	domain := "foo.example.org"
	req, _ := http.NewRequest("GET", "/update?secret=changeme&domain="+domain+"&addr=1.2.3.4", nil)
	verify(t, req, e, domain, RecordUpdateRequest{
		zone:        "",
		fqdn:        "foo.example.org",
		ddnsKeyName: "foo.example.org",
	})
}

func TestExtractorUnconfiguredZoneWithZoneInRequest(t *testing.T) {
	var e = defaultRequestDataExtractor{appConfig: &Config{
		Zone: "",
	}}

	domain := "foo"
	req, _ := http.NewRequest("GET", "/update?secret=changeme&domain="+domain+"&addr=1.2.3.4&zone=example.org", nil)
	verify(t, req, e, domain, RecordUpdateRequest{
		zone:        "example.org",
		fqdn:        "foo.example.org",
		ddnsKeyName: "example.org",
	})
}

func TestExtractorUnconfiguredZoneWithDDnskeyInRequest(t *testing.T) {
	var e = defaultRequestDataExtractor{appConfig: &Config{
		Zone: "",
	}}

	domain := "foo.example.org"
	req, _ := http.NewRequest("GET", "/update?secret=changeme&domain="+domain+"&addr=1.2.3.4&ddnskeyname=example.org", nil)
	verify(t, req, e, domain, RecordUpdateRequest{
		zone:        "",
		fqdn:        "foo.example.org",
		ddnsKeyName: "example.org",
	})
}

func TestExtractorConfiguredZoneAndOnlyWithHostname(t *testing.T) {
	var e = defaultRequestDataExtractor{appConfig: &Config{
		Zone: "example.org.",
	}}

	domain := "foo"
	req, _ := http.NewRequest("GET", "/update?secret=changeme&domain="+domain+"&addr=1.2.3.4", nil)
	verify(t, req, e, domain, RecordUpdateRequest{
		zone:        "example.org",
		fqdn:        "foo.example.org",
		ddnsKeyName: "example.org",
	})
}

func TestExtractorConfiguredZoneAndOnlyWithFQDN(t *testing.T) {
	var e = defaultRequestDataExtractor{appConfig: &Config{
		Zone: "example.org.",
	}}

	domain := "foo.example.org."
	req, _ := http.NewRequest("GET", "/update?secret=changeme&domain="+domain+"&addr=1.2.3.4", nil)
	verify(t, req, e, domain, RecordUpdateRequest{
		zone:        "",
		fqdn:        "foo.example.org",
		ddnsKeyName: "foo.example.org",
	})
}
