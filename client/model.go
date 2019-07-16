package client

import (
	"net"
	"syscall"
)

type Model struct {
	Nickname       string
	Port           int
	socket         int
	token          string
	sockAddrRemote syscall.Sockaddr
}

func New(nickname string) *Model {
	return &Model{
		Nickname: nickname,
	}
}

func (client *Model) Connect(addr string, port int) error {
	if err := client.createTCPSocket(); err != nil {
		return err
	}

	if err := client.dial(addr, port); err != nil {
		return err
	}

	err := client.Send("hallo")
	return err
}

func (client *Model) Send(message string) error {
	return syscall.Sendto(client.socket, []byte(message), 0, client.sockAddrRemote)
}

func (client *Model) createTCPSocket() error {
	fd, err := syscall.Socket(syscall.AF_INET6, syscall.SOCK_STREAM, 0)
	if err != nil {
		return err
	}
	client.socket = fd
	return nil
}

func (client *Model) dial(addr string, port int) error {
	ipAddr, err := net.ResolveIPAddr("", addr)
	if err != nil {
		return err
	}

	// check if ipv6
	if ipAddr.IP.To4() == nil {
		ipv6, err := convertIPv6ToArray(ipAddr)
		if err != nil {
			return err
		}
		client.sockAddrRemote = &syscall.SockaddrInet6{
			Port: port,
			Addr: ipv6,
		}
	} else {
		ipv4, err := convertIPv4ToArray(ipAddr)
		if err != nil {
			return err
		}
		client.sockAddrRemote = &syscall.SockaddrInet4{
			Port: port,
			Addr: ipv4,
		}
	}
	return nil
}