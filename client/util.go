package client

import (
	"errors"
	"fmt"
	"math/rand"
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

func getRandomUserPort() int {
	size := 0xbfff - 0x400
	return rand.Intn(size) + 0x400
}

//ToNetstring converts a string to netstring
func ToNetstring(message string) string {
	return fmt.Sprintf("%v:%v,", len(message), message)
}

//FromNetstring converts a netstring to a string
func FromNetstring(message string) (string, error) {
	lst := strings.SplitN(message, ":", 2)
	if len(lst) < 2 {
		return message, fmt.Errorf("not a netstring: %v", message)
	}
	length, err := strconv.Atoi(lst[0])
	if err != nil {
		return message, fmt.Errorf("not a netstring: %v", message)
	}
	return lst[1][:length], nil
}

