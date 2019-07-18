package client

import (
	"fmt"
	"github.com/golang/glog"
	"golang.org/x/sys/unix"
	"net"
	"strconv"
	"time"
)

const (
	CtrlClientPrefix   = "C"
	CtrlServerPrefix   = "S"
	CtrlReceiveTimeout = 3 * time.Second
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

	glog.Info("successfully exchanged protocol version")
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

//sendCtrl is a helper method for sending strings over the control channel
func (client *Model) sendCtrl(message string) error {
	errMsg := "[control] send error: %v"
	netstring := ToNetstring(fmt.Sprintf("%v %v", CtrlClientPrefix, message))
	err := unix.Sendto(client.ctrlSocket, []byte(netstring), 0, client.sockAddrRemote)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	return nil
}

func (client *Model) sendCtrlError(message string) error {
	errMsg := "[control] send error: %v"
	netstring := ToNetstring(fmt.Sprintf("%v %v", ErrorPrefix, message))
	err := unix.Sendto(client.ctrlSocket, []byte(netstring), 0, client.sockAddrRemote)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	return nil
}

//recvCtrl is a helper method for receiving strings over the control channel
func (client *Model) recvCtrl() (string, error) {
	errMsg := "[control] %v"
	msg, err := client.recv(client.ctrlSocket, CtrlReceiveTimeout, CtrlServerPrefix)
	if err != nil {
		_ = client.sendCtrlError(err.Error())
		return msg, fmt.Errorf(errMsg, err)
	}
	return msg, nil
}

//createCtrlSocket creates a socket for the control channel
func (client *Model) createCtrlSocket() error {
	fd, err := unix.Socket(unix.AF_INET6, unix.SOCK_STREAM, 0)
	if err != nil {
		return err
	}
	client.ctrlSocket = fd
	return nil
}

// dial creates a connection to the remote host
func (client *Model) dial() error {
	errMsg := "[control] dial error: %v"

	// create a tcp socket
	if err := client.createCtrlSocket(); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	// get ip address from url
	ipAddr, err := net.ResolveIPAddr("ip6", client.remoteIP)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	ipv6, err := convertIPv6ToArray(ipAddr)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	client.sockAddrRemote = &unix.SockaddrInet6{
		Port: client.remotePort,
		Addr: ipv6,
	}

	err = unix.Connect(client.ctrlSocket, client.sockAddrRemote)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	glog.Info("established connection to remote server on the control channel")
	return nil
}

func (client *Model) validateStringLength() error {
	errMsg := "[control] validate string length: %v"
	lenStr, err := client.recvCtrl()
	if err != nil{
		return fmt.Errorf(errMsg, err)
	}
	l, err := strconv.Atoi(lenStr)
	if err != nil{
		return fmt.Errorf(errMsg, err)
	}
	if l != client.lastMessageLen {
		err := fmt.Errorf("string lengths do not match: server=%v client=%v", l, client.lastMessageLen)
		return fmt.Errorf(errMsg, err)
	}
	glog.Info("server sent valid string length")
	return nil
}