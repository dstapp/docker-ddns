package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"dyndns/ipparser"
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

	if ipparser.ValidIP4(response.Address) {
		response.AddrType = "A"
	} else if ipparser.ValidIP6(response.Address) {
		response.AddrType = "AAAA"
	} else {
		var ip string
		var err error

		ip, err = getUserIP(r)
		if ip == "" {
			ip, _, err = net.SplitHostPort(r.RemoteAddr)
		}

		if err != nil {
			response.Success = false
			response.Message = fmt.Sprintf("%q is neither a valid IPv4 nor IPv6 address", r.RemoteAddr)
			log.Println(fmt.Sprintf("Invalid address: %q", r.RemoteAddr))
			return response
		}

		// @todo refactor this code to remove duplication
		if ipparser.ValidIP4(ip) {
			response.AddrType = "A"
		} else if ipparser.ValidIP6(ip) {
			response.AddrType = "AAAA"
		} else {
			response.Success = false
			response.Message = fmt.Sprintf("%s is neither a valid IPv4 nor IPv6 address", response.Address)
			log.Println(fmt.Sprintf("Invalid address: %s", response.Address))
			return response
		}

		response.Address = ip
	}

	response.Success = true

	return response
}

func getUserIP(r *http.Request) (string, error) {
	for _, h := range []string{"X-Real-Ip", "X-Forwarded-For"} {
		addresses := strings.Split(r.Header.Get(h), ",")
		// march from right to left until we get a public address
		// that will be the address right before our proxy.
		for i := len(addresses) - 1; i >= 0; i-- {
			ip := strings.TrimSpace(addresses[i])
			// header can contain spaces too, strip those out.
			realIP := net.ParseIP(ip)
			if !realIP.IsGlobalUnicast() || isPrivateSubnet(realIP) {
				// bad address, go to next
				continue
			}
			return ip, nil
		}
	}
	return "", errors.New("no match")
}

//ipRange - a structure that holds the start and end of a range of ip addresses
type ipRange struct {
	start net.IP
	end   net.IP
}

// inRange - check to see if a given ip address is within a range given
func inRange(r ipRange, ipAddress net.IP) bool {
	// strcmp type byte comparison
	if bytes.Compare(ipAddress, r.start) >= 0 && bytes.Compare(ipAddress, r.end) < 0 {
		return true
	}
	return false
}

var privateRanges = []ipRange{
	ipRange{
		start: net.ParseIP("10.0.0.0"),
		end:   net.ParseIP("10.255.255.255"),
	},
	ipRange{
		start: net.ParseIP("100.64.0.0"),
		end:   net.ParseIP("100.127.255.255"),
	},
	ipRange{
		start: net.ParseIP("172.16.0.0"),
		end:   net.ParseIP("172.31.255.255"),
	},
	ipRange{
		start: net.ParseIP("192.0.0.0"),
		end:   net.ParseIP("192.0.0.255"),
	},
	ipRange{
		start: net.ParseIP("192.168.0.0"),
		end:   net.ParseIP("192.168.255.255"),
	},
	ipRange{
		start: net.ParseIP("198.18.0.0"),
		end:   net.ParseIP("198.19.255.255"),
	},
}

// isPrivateSubnet - check to see if this ip is in a private subnet
func isPrivateSubnet(ipAddress net.IP) bool {
	// my use case is only concerned with ipv4 atm
	if ipCheck := ipAddress.To4(); ipCheck != nil {
		// iterate over all our ranges
		for _, r := range privateRanges {
			// check if this ip is in a private range
			if inRange(r, ipAddress) {
				return true
			}
		}
	}
	return false
}

