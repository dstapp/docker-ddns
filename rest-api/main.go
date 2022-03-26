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
	router.Handle("/update", requestRequestDataMiddleware(http.HandlerFunc(update), defaultExtractor)).Methods(http.MethodGet)
	router.Handle("/delete", requestRequestDataMiddleware(http.HandlerFunc(update), defaultExtractor)).Methods(http.MethodGet, http.MethodDelete)

	/* DynDNS compatible handlers. Most routers will invoke /nic/update */
	router.Handle("/nic/update", requestRequestDataMiddleware(http.HandlerFunc(dynUpdate), dynExtractor)).Methods(http.MethodGet)
	router.Handle("/v2/update", requestRequestDataMiddleware(http.HandlerFunc(dynUpdate), dynExtractor)).Methods(http.MethodGet)
	router.Handle("/v3/update", requestRequestDataMiddleware(http.HandlerFunc(dynUpdate), dynExtractor)).Methods(http.MethodGet)

	listenTo := fmt.Sprintf("%s:%d", "", appConfig.Port)

	log.Println(fmt.Sprintf("Serving dyndns REST services on " + listenTo + "..."))
	log.Fatal(http.ListenAndServe(listenTo, router))
}

func dynUpdate(w http.ResponseWriter, r *http.Request) {
	response := r.Context().Value(responseKey).(WebserviceResponse)

	if !response.Success {
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

	if !response.Success {
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

	for _, record := range response.Records {
		for _, domain := range response.Domains {
			recordUpdate := RecordUpdateRequest{
				domain:      domain,
				ipAddr:      record.Value,
				addrType:    record.Type,
				secret:      extractor.Secret(r),
				ddnsKeyName: extractor.DdnsKeyName(r, domain),
				zone:        extractor.Zone(r, domain),
				fqdn:        extractor.Fqdn(r, domain),
				action:      extractor.Action(r),
			}
			result, err := recordUpdate.updateRecord()

			if err != nil {
				response.Success = false
				response.Message = err.Error()

				onError()
				return false
			}
			response.Success = true
			if len(response.Message) != 0 {
				response.Message += "; "
			}
			response.Message += result
		}
	}

	return true
}

func (r RecordUpdateRequest) updateRecord() (string, error) {
	var nsupdate NSUpdateInterface = NewNSUpdate()
	message := "No action executed"
	switch r.action {
	case UpdateRequestActionDelete:
		nsupdate.DeleteRecord(r)
		message = fmt.Sprintf("Deleted %s record for %s", r.addrType, r.domain)
	case UpdateRequestActionUpdate:
		fallthrough
	default:
		nsupdate.UpdateRecord(r)
		message = fmt.Sprintf("Updated %s record: %s -> %s", r.addrType, r.domain, r.ipAddr)
	}
	result := nsupdate.Close()

	log.Println(message)

	if result != "" {
		return "", fmt.Errorf("%s", result)
	}
	return message, nil
}
