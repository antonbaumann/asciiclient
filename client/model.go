package client

import (
	"fmt"
	"net"
	"syscall"
)

const (
	ClientCtrlPrefix = "C"
	ServerCtrlPrefix = "S"
	Protocol         = "GRNVS V:1.0"
)

type Model struct {
	Nickname       string
	Port           int
	ctrlSocket     int
	dataSocket     int
	token          string
	sockAddrRemote syscall.Sockaddr
	buffer         []byte
}

func New(nickname string) *Model {
	return &Model{
		Nickname: nickname,
		buffer: make([]byte, 2000),
	}
}

func (client *Model) Connect(addr string, port int) error {
	errMsg := "connect error: %v"

	if err := client.dial(addr, port); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	protocolMsg := fmt.Sprintf("%v %v", ClientCtrlPrefix, Protocol)
	if err := client.sendCtrl(ToNetstring(protocolMsg)); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	msg, err := client.recvCtrl()
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	fmt.Println(msg)

	return nil
}

func (client *Model) sendCtrl(message string) error {
	errMsg := "[control] send error: %v"
	err := syscall.Sendto(client.ctrlSocket, []byte(message), 0, client.sockAddrRemote)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	return nil
}

func (client *Model) recvCtrl() (string, error) {
	errMsg := "[control] receive error: %v"
	length, _, err := syscall.Recvfrom(client.ctrlSocket, client.buffer, 0)
	if err != nil {
		return "", fmt.Errorf(errMsg, err)
	}
	data := client.buffer[:length]
	message, err := FromNetstring(string(data))
	if err != nil {
		return message, fmt.Errorf(errMsg, err)
	}
	return message, nil
}

func (client *Model) createCtrlSocket() error {
	fd, err := syscall.Socket(syscall.AF_INET6, syscall.SOCK_STREAM, 0)
	if err != nil {
		return err
	}
	client.ctrlSocket = fd
	return nil
}

func (client *Model) dial(addr string, port int) error {
	errMsg := "[control] dial error: %v"

	// create a tcp socket
	if err := client.createCtrlSocket(); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	// get ip address from url
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

	err = syscall.Connect(client.ctrlSocket, client.sockAddrRemote)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	return nil
}
