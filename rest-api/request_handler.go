package main

import (
    "log"
    "fmt"
    "net/http"

    "dyndns/ipparser"
)

type WebserviceResponse struct {
    Success bool
    Message string
    Domain string
    Address string
    AddrType string
}

func BuildWebserviceResponseFromRequest(r *http.Request, appConfig *Config) WebserviceResponse {
    response := WebserviceResponse{}

    var sharedSecret string

    vals := r.URL.Query()
    sharedSecret = vals.Get("secret")
    response.Domain = vals.Get("domain")
    response.Address = vals.Get("addr")

    if sharedSecret != appConfig.SharedSecret {
        log.Println(fmt.Sprintf("Invalid shared secret: %s", sharedSecret))
        response.Success = false
        response.Message = "Invalid Credentials"
        return response
    }

    if response.Domain == "" {
        response.Success = false
        response.Message = fmt.Sprintf("Domain not set")
        log.Println("Domain not set")
        return response
    }

    if ipparser.ValidIP4(response.Address) {
        response.AddrType = "A"
    } else if ipparser.ValidIP6(response.Address) {
        response.AddrType = "AAAA"
    } else {
        response.Success = false
        response.Message = fmt.Sprintf("%s is neither a valid IPv4 nor IPv6 address", response.Address)
        log.Println(fmt.Sprintf("Invalid address: %s", response.Address))
        return response
    }

    response.Success = true

    return response
}
