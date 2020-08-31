package main

import (
	"context"
	"net/http"
	"strings"
)

type requestDataExtractor interface {
	Address(r *http.Request) string
	Secret(r *http.Request) string
	Domain(r *http.Request) string
	DdnsKeyName(r *http.Request, domain string) string
	Zone(r *http.Request, domain string) string
	Fqdn(r *http.Request, domain string) string
}

type defaultRequestDataExtractor struct {
	appConfig *Config
}

func (e defaultRequestDataExtractor) Address(r *http.Request) string {
	return r.URL.Query().Get("addr")
}
func (e defaultRequestDataExtractor) Secret(r *http.Request) string {
	return r.URL.Query().Get("secret")
}
func (e defaultRequestDataExtractor) Domain(r *http.Request) string {
	return r.URL.Query().Get("domain")
}
func (e defaultRequestDataExtractor) DdnsKeyName(r *http.Request, domain string) string {
	ddnsKeyName := r.URL.Query().Get("ddnskeyname")
	if ddnsKeyName != "" {
		return ddnsKeyName
	}
	ddnsKeyName = e.Zone(r, domain)
	if ddnsKeyName != "" {
		return ddnsKeyName
	}
	ddnsKeyName = e.Fqdn(r, domain)
	return ddnsKeyName
}
func (e defaultRequestDataExtractor) Zone(r *http.Request, domain string) string {
	zone := r.URL.Query().Get("zone")
	if zone != "" {
		return zone
	}
	zone = strings.TrimRight(e.appConfig.Zone, ".")
	if domain[len(domain)-1:] == "." {
		zone = ""
	}
	return zone
}
func (e defaultRequestDataExtractor) Fqdn(r *http.Request, domain string) string {
	return strings.TrimRight(escape(domain)+"."+e.Zone(r, domain), ".")
}

type dynRequestDataExtractor struct{ defaultRequestDataExtractor }

func (e dynRequestDataExtractor) Secret(r *http.Request) string {
	_, sharedSecret, ok := r.BasicAuth()
	if !ok || sharedSecret == "" {
		sharedSecret = r.URL.Query().Get("password")
	}

	return sharedSecret
}
func (e dynRequestDataExtractor) Address(r *http.Request) string {
	return r.URL.Query().Get("myip")
}
func (e dynRequestDataExtractor) Domain(r *http.Request) string {
	return r.URL.Query().Get("hostname")
}

func requestRequestDataMiddleware(next http.Handler, extractors requestDataExtractor) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), extractorKey, extractors)
		response := BuildWebserviceResponseFromRequest(r, &appConfig.Config, extractors)
		ctx = context.WithValue(ctx, responseKey, response)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
