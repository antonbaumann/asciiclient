package client

import (
	"errors"
	"fmt"
	"net"
)

func convertIPv4ToArray(addr *net.IPAddr) ([4]byte, error) {
	ipv4 := addr.IP.To4()
	var arr [4]byte

	if ipv4 == nil {
		return arr, errors.New("invalid IPv4 address")
	}
	copy(arr[:], ipv4)
	return arr, nil
}

func convertIPv6ToArray(addr *net.IPAddr) ([16]byte, error) {
	ipv6 := addr.IP.To16()
	var arr [16]byte

	if ipv6 == nil {
		return arr, errors.New("invalid IPv6 address")
	}
	copy(arr[:], ipv6)
	return arr, nil
}

func Netstring(message string) string {
	return fmt.Sprintf("%v:%v,", len(message), message)
}

