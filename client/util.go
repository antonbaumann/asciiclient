package client

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
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

func ToNetstring(message string) string {
	return fmt.Sprintf("%v:%v,", len(message), message)
}

func FromNetstring(message string) (string, error) {
	lst := strings.SplitN(message, ":", 1)
	if len(lst) < 2 {
		return message, fmt.Errorf("not a netstring: %v", message)
	}
	length, err := strconv.Atoi(lst[0])
	if err != nil {
		return message, fmt.Errorf("not a netstring: %v", message)
	}
	return lst[1][:length], nil
}

