package main

import (
    "log"
    "fmt"
    "net"
    "net/http"
    "strings"

    "dyndns/ipparser"
)

type WebserviceResponse struct {
    Success bool
    Message string
    Domain string
    Domains []string
    Address string
    AddrType string
}

func BuildWebserviceResponseFromRequest(r *http.Request, appConfig *Config) WebserviceResponse {
    response := WebserviceResponse{}

    var sharedSecret string

    vals := r.URL.Query()
    sharedSecret = vals.Get("secret")
    response.Domains = strings.Split(vals.Get("domain"), ",")
    response.Address = vals.Get("addr")

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

    if ipparser.ValidIP4(response.Address) {
        response.AddrType = "A"
    } else if ipparser.ValidIP6(response.Address) {
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
        if ipparser.ValidIP4(ip) {
            response.AddrType = "A"
            response.Address = ip
        } else if ipparser.ValidIP6(ip) {
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

func BuildDynResponseFromRequest(r *http.Request, appConfig *Config) WebserviceResponse {
    response := WebserviceResponse{}

    var sharedSecret string

    vals := r.URL.Query()
    _, sharedSecret, ok := r.BasicAuth()
    if !ok || sharedSecret == "" {
        sharedSecret = vals.Get("password")
    }

    response.Domains = strings.Split(vals.Get("hostname"), ",")
    response.Address = vals.Get("myip")

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

    if ipparser.ValidIP4(response.Address) {
        response.AddrType = "A"
    } else if ipparser.ValidIP6(response.Address) {
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
        if ipparser.ValidIP4(ip) {
            response.AddrType = "A"
            response.Address = ip
        } else if ipparser.ValidIP6(ip) {
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
