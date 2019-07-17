package client

import (
	"fmt"
	"net"
	"strings"
	"syscall"
)

const (
	CtrlClientPrefix = "C"
	CtrlServerPrefix = "S"
)

//exchangeProtocolVersion initiates the connection and checks
// if the protocol version of client and server matches
func (client *Model) exchangeProtocolVersion() error {
	errMsg := "[control] protocol exchange error: %v"

	if err := client.sendProtocolVersion(); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	if err := client.receiveProtocolVersion(); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	return nil
}

//sendProtocolVersion sends the protocol version used by the client to the server
func (client *Model) sendProtocolVersion() error {
	if err := client.sendCtrl(ProtocolVersion); err != nil {
		return err
	}
	return nil
}

//receiveProtocolVersion receives the protocol version used by the server
//it returns an error if versions do not match
func (client *Model) receiveProtocolVersion() error {
	msg, err := client.recvCtrl()
	if err != nil {
		return err
	}
	if msg != ProtocolVersion {
		return err
	}
	return nil
}

//sendNickname sends the client nickname to the server
func (client *Model) sendNickname() error {
	errMsg := "[control] send nickname error: %v"
	if err := client.sendCtrl(client.Nickname); err != nil {
		return fmt.Errorf(errMsg, err)
	}
	return nil
}

//recvToken receives the token from the server
func (client *Model) recvToken() error {
	errMsg := "[control] receive token error: %v"
	token, err := client.recvCtrl()
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	client.token = token

	return nil
}

//sendCtrl is a helper method for sending strings over the control channel
func (client *Model) sendCtrl(message string) error {
	errMsg := "[control] send error: %v"
	netstring := ToNetstring(fmt.Sprintf("%v %v", CtrlClientPrefix, message))
	err := syscall.Sendto(client.ctrlSocket, []byte(netstring), 0, client.sockAddrRemote)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	return nil
}

//recvCtrl is a helper method for receiving strings over the control channel
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
	return parseServerMessage(message)
}

func parseServerMessage(message string) (string, error) {
	lst := strings.SplitN(message, " ", 2)
	if len(lst) < 2 {
		return "", fmt.Errorf("server response malformed: %v", message)
	}
	if lst[0] == ErrorPrefix {
		return "", fmt.Errorf("server error: %v", message)
	}
	if lst[0] == CtrlServerPrefix {
		return lst[1], nil
	}
	return "", fmt.Errorf("server response malformed: %v", message)
}

//createCtrlSocket creates a socket for the control channel
func (client *Model) createCtrlSocket() error {
	fd, err := syscall.Socket(syscall.AF_INET6, syscall.SOCK_STREAM, 0)
	if err != nil {
		return err
	}
	client.ctrlSocket = fd
	return nil
}

// dial creates a connection to the remote host
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

