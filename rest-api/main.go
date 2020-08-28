package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var appConfig = &ConfigFlags{}

func main() {
	appConfig.LoadConfig()

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/update", Update).Methods("GET")

	/* DynDNS compatible handlers. Most routers will invoke /nic/update */
	router.HandleFunc("/nic/update", DynUpdate).Methods("GET")
	router.HandleFunc("/v2/update", DynUpdate).Methods("GET")
	router.HandleFunc("/v3/update", DynUpdate).Methods("GET")

	listenTo := fmt.Sprintf("%s:%d", "", appConfig.Port)

	log.Println(fmt.Sprintf("Serving dyndns REST services on " + listenTo + "..."))
	log.Fatal(http.ListenAndServe(listenTo, router))
}

func DynUpdate(w http.ResponseWriter, r *http.Request) {
	extractor := RequestDataExtractor{
		Address: func(r *http.Request) string { return r.URL.Query().Get("myip") },
		Secret: func(r *http.Request) string {
			_, sharedSecret, ok := r.BasicAuth()
			if !ok || sharedSecret == "" {
				sharedSecret = r.URL.Query().Get("password")
			}

			return sharedSecret
		},
		Domain: func(r *http.Request) string { return r.URL.Query().Get("hostname") },
	}
	response := BuildWebserviceResponseFromRequest(r, &appConfig.Config, extractor)

	if response.Success == false {
		if response.Message == "Domain not set" {
			w.Write([]byte("notfqdn\n"))
		} else {
			w.Write([]byte("badauth\n"))
		}
		return
	}

	for _, domain := range response.Domains {
		recordUpdate := RecordUpdateRequest{
			domain:   domain,
			ipaddr:   response.Address,
			addrType: response.AddrType,
			ddnskey:  "",
		}
		result := recordUpdate.updateRecord()

		if result != "" {
			response.Success = false
			response.Message = result

			w.Write([]byte("dnserr\n"))
			return
		}
	}

	response.Success = true
	response.Message = fmt.Sprintf("Updated %s record for %s to IP address %s", response.AddrType, response.Domain, response.Address)

	w.Write([]byte(fmt.Sprintf("good %s\n", response.Address)))
}

func Update(w http.ResponseWriter, r *http.Request) {
	extractor := RequestDataExtractor{
		Address: func(r *http.Request) string { return r.URL.Query().Get("addr") },
		Secret:  func(r *http.Request) string { return r.URL.Query().Get("secret") },
		Domain:  func(r *http.Request) string { return r.URL.Query().Get("domain") },
	}
	response := BuildWebserviceResponseFromRequest(r, &appConfig.Config, extractor)

	if response.Success == false {
		json.NewEncoder(w).Encode(response)
		return
	}

	for _, domain := range response.Domains {
		recordUpdate := RecordUpdateRequest{
			domain:   domain,
			ipaddr:   response.Address,
			addrType: response.AddrType,
			ddnskey:  "",
		}
		result := recordUpdate.updateRecord()

		if result != "" {
			response.Success = false
			response.Message = result

			json.NewEncoder(w).Encode(response)
			return
		}
	}

	response.Success = true
	response.Message = fmt.Sprintf("Updated %s record for %s to IP address %s", response.AddrType, response.Domain, response.Address)

	json.NewEncoder(w).Encode(response)
}

func (r RecordUpdateRequest) updateRecord() string {
	var nsupdate NSUpdateInterface = NewNSUpdate()
	nsupdate.UpdateRecord(r)
	result := nsupdate.Close()

	status := "succeeded"
	if result != "" {
		status = "failed, error: " + result
	}

	log.Println(fmt.Sprintf("%s record update request: %s -> %s %s", r.addrType, r.domain, r.ipaddr, status))

	return result
}
