package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
)

type RequestDataExtractor struct {
	Address func(request *http.Request) string
	Secret  func(request *http.Request) string
	Domain  func(request *http.Request) string
}

type WebserviceResponse struct {
	Success  bool
	Message  string
	Domain   string
	Domains  []string
	Address  string
	AddrType string
}

func BuildWebserviceResponseFromRequest(r *http.Request, appConfig *Config, extractors RequestDataExtractor) WebserviceResponse {
	response := WebserviceResponse{}

	sharedSecret := extractors.Secret(r)
	response.Domains = strings.Split(extractors.Domain(r), ",")
	response.Address = extractors.Address(r)

	if sharedSecret != appConfig.SharedSecret {
		log.Println(fmt.Sprintf("Invalid shared secret: %s", sharedSecret))
		response.Success = false
		response.Message = "Invalid Credentials"
		return response
	}

	for _, domain := range response.Domains {
		if domain == "" {
			response.Success = false
			response.Message = fmt.Sprintf("Domain not set")
			log.Println("Domain not set")
			return response
		}
	}

	// kept in the response for compatibility reasons
	response.Domain = strings.Join(response.Domains, ",")

	if ValidIP4(response.Address) {
		response.AddrType = "A"
	} else if ValidIP6(response.Address) {
		response.AddrType = "AAAA"
	} else {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)

		if err != nil {
			response.Success = false
			response.Message = fmt.Sprintf("%q is neither a valid IPv4 nor IPv6 address", r.RemoteAddr)
			log.Println(fmt.Sprintf("Invalid address: %q", r.RemoteAddr))
			return response
		}

		// @todo refactor this code to remove duplication
		if ValidIP4(ip) {
			response.AddrType = "A"
			response.Address = ip
		} else if ValidIP6(ip) {
			response.AddrType = "AAAA"
			response.Address = ip
		} else {
			response.Success = false
			response.Message = fmt.Sprintf("%s is neither a valid IPv4 nor IPv6 address", response.Address)
			log.Println(fmt.Sprintf("Invalid address: %s", response.Address))
			return response
		}
	}

	response.Success = true

	return response
}
