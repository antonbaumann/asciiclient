package client

import (
	"fmt"
	"github.com/golang/glog"
	"strconv"
	"syscall"
	"time"
)

const (
	DataClientPrefix   = "D"
	DataServerPrefix   = "T"
	DataReceiveTimeout = 5 * time.Second
)

//createListeningSocket creates a socket for the data channel
func (client *Model) createListeningSocket() error {
	errMsg := "[data] create data socket error: %v"
	fd, err := syscall.Socket(syscall.AF_INET6, syscall.SOCK_STREAM, 0)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	port, err := bindRandomPort(fd)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	client.ListeningPort = port
	client.dataSocket = fd

	if err := client.sendListeningPortNumber(); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	return nil
}

func (client *Model) sendListeningPortNumber() error {
	errMsg := "[data] send data port error: %v"
	if err := client.sendCtrl(strconv.Itoa(client.ListeningPort)); err != nil {
		return fmt.Errorf(errMsg, err)
	}
	glog.Infof("sent listening port [%v] to server", client.ListeningPort)
	glog.Infof("listening now on port %v", client.ListeningPort)
	return nil
}

func bindRandomPort(socket int) (int, error) {
	errMsg := "[data] bind port to data socket error: %v"
	var inaddr6Any [16]byte
	var err error
	nrRetries := 5
	for i := 0; i < nrRetries; i++ {
		port := GetRandomUserPort()
		sa := &syscall.SockaddrInet6{
			Port: port,
			Addr: inaddr6Any,
		}
		if err = syscall.Bind(socket, sa); err == nil {
			return port, nil
		}
	}
	return -1, fmt.Errorf(errMsg, err)
}

func (client *Model) recvProtocolConfirmation(fd int, sa syscall.Sockaddr) error {
	errMsg := "receive protocol confirmation error: %v"
	msg, err := client.recvData(fd)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	if msg != ProtocolVersion {
		return fmt.Errorf(errMsg, "protocols do not match")
	}
	return nil
}

func (client *Model) recvData(fd int) (string, error) {
	errMsg := "[data] %v"
	msg, err := client.recv(fd, DataReceiveTimeout, DataServerPrefix)
	if err != nil {
		return msg, fmt.Errorf(errMsg, err)
	}
	return msg, nil
}


func (client *Model) awaitServerConnection() (int, syscall.Sockaddr, error) {
	// todo timeout
	errMsg := "[ctrl] await server connection: %v"
	timeout := SecondsToTimeval(int(DataReceiveTimeout.Seconds()))
	_, err := syscall.Select(0xbfff +1, SocketToFDSet(client.dataSocket), nil, nil, timeout)
	if err != nil {
		return -1, nil, fmt.Errorf(errMsg, err)
	}
	nfd, sa, err := syscall.Accept(client.dataSocket)
	if err != nil {
		return -1, nil, fmt.Errorf(errMsg, err)
	}
	return nfd, sa, nil
}