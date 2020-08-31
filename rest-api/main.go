package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type key int

const (
	responseKey  key = iota
	extractorKey key = iota
)

var appConfig = &ConfigFlags{}

func main() {
	defaultExtractor := defaultRequestDataExtractor{appConfig: &appConfig.Config}
	dynExtractor := dynRequestDataExtractor{defaultRequestDataExtractor{appConfig: &appConfig.Config}}

	appConfig.LoadConfig()

	router := mux.NewRouter().StrictSlash(true)
	router.Handle("/update", requestRequestDataMiddleware(http.HandlerFunc(update), defaultExtractor)).Methods("GET")

	/* DynDNS compatible handlers. Most routers will invoke /nic/update */
	router.Handle("/nic/update", requestRequestDataMiddleware(http.HandlerFunc(dynUpdate), dynExtractor)).Methods("GET")
	router.Handle("/v2/update", requestRequestDataMiddleware(http.HandlerFunc(dynUpdate), dynExtractor)).Methods("GET")
	router.Handle("/v3/update", requestRequestDataMiddleware(http.HandlerFunc(dynUpdate), dynExtractor)).Methods("GET")

	listenTo := fmt.Sprintf("%s:%d", "", appConfig.Port)

	log.Println(fmt.Sprintf("Serving dyndns REST services on " + listenTo + "..."))
	log.Fatal(http.ListenAndServe(listenTo, router))
}

func dynUpdate(w http.ResponseWriter, r *http.Request) {
	response := r.Context().Value(responseKey).(WebserviceResponse)

	if response.Success == false {
		if response.Message == "Domain not set" {
			w.Write([]byte("notfqdn\n"))
		} else {
			w.Write([]byte("badauth\n"))
		}
		return
	}

	success := updateDomains(r, &response, func() {
		w.Write([]byte("dnserr\n"))
	})

	if !success {
		return
	}

	w.Write([]byte(fmt.Sprintf("good %s\n", response.Address)))
}

func update(w http.ResponseWriter, r *http.Request) {
	response := r.Context().Value(responseKey).(WebserviceResponse)

	if response.Success == false {
		json.NewEncoder(w).Encode(response)
		return
	}

	success := updateDomains(r, &response, func() {
		json.NewEncoder(w).Encode(response)
	})

	if !success {
		return
	}

	json.NewEncoder(w).Encode(response)
}

func updateDomains(r *http.Request, response *WebserviceResponse, onError func()) bool {
	extractor := r.Context().Value(extractorKey).(requestDataExtractor)

	for _, address := range response.Addresses {
		for _, domain := range response.Domains {
			recordUpdate := RecordUpdateRequest{
				domain:      domain,
				ipAddr:      address.Address,
				addrType:    address.AddrType,
				secret:      extractor.Secret(r),
				ddnsKeyName: extractor.DdnsKeyName(r, domain),
				zone:        extractor.Zone(r, domain),
				fqdn:        extractor.Fqdn(r, domain),
			}
			result := recordUpdate.updateRecord()

			if result != "" {
				response.Success = false
				response.Message = result

				onError()
				return false
			}

			response.Success = true
			if len(response.Message) != 0 {
				response.Message += "; "
			}
			response.Message += fmt.Sprintf("Updated %s record for %s to IP address %s", address.AddrType, domain, address.Address)
		}
	}

	return true
}

func (r RecordUpdateRequest) updateRecord() string {
	var nsupdate NSUpdateInterface = NewNSUpdate()
	nsupdate.UpdateRecord(r)
	result := nsupdate.Close()

	status := "succeeded"
	if result != "" {
		status = "failed, error: " + result
	}

	log.Println(fmt.Sprintf("%s record update request: %s -> %s %s", r.addrType, r.domain, r.ipAddr, status))

	return result
}
