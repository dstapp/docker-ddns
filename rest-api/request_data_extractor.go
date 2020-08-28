package main

import (
	"context"
	"net/http"
)

type requestDataExtractor interface {
	Address(r *http.Request) string
	Secret(r *http.Request) string
	Domain(r *http.Request) string
	DdnsKey(r *http.Request) string
}

type defaultRequestDataExtractor struct{}

func (e defaultRequestDataExtractor) Address(r *http.Request) string {
	return r.URL.Query().Get("addr")
}
func (e defaultRequestDataExtractor) Secret(r *http.Request) string {
	return r.URL.Query().Get("secret")
}
func (e defaultRequestDataExtractor) Domain(r *http.Request) string {
	return r.URL.Query().Get("domain")
}
func (e defaultRequestDataExtractor) DdnsKey(r *http.Request) string {
	return r.URL.Query().Get("ddnskey")
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
