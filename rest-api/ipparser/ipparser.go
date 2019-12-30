package ipparser

import (
	"net"
)

func ValidIP4(ipAddress string) bool {
	testInput := net.ParseIP(ipAddress)
	if testInput == nil {
		return false
	}

	return (testInput.To4() != nil)
}

func ValidIP6(ip6Address string) bool {
	testInputIP6 := net.ParseIP(ip6Address)
	if testInputIP6 == nil {
		return false
	}

	return (testInputIP6.To16() != nil)
}
