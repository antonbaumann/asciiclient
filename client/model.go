package client

import (
	"fmt"
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

	err := client.Send("C GRNVS V:1.0")
	return err
}

func (client *Model) Send(message string) error {
	errMsg := "send error: %v"
	err := syscall.Sendto(client.socket, []byte(message), 0, client.sockAddrRemote)
	return fmt.Errorf(errMsg, err)
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
	errMsg := "dial error: %v"

	ipAddr, err := net.ResolveIPAddr("ip6", addr)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	ipv6, err := convertIPv6ToArray(ipAddr)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	client.sockAddrRemote = &syscall.SockaddrInet6{
		Port: port,
		Addr: ipv6,
	}

	err = syscall.Connect(client.socket, client.sockAddrRemote)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	return nil
}
