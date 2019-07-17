package client

import (
	"fmt"
	"syscall"
)

const (
	DataClientPrefix = "D"
	DataServerPrefix = "T"
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

	return nil
}

func (client *Model) sendListeningPortNumber() error {
	errMsg := "[data] send data port error: %v"
	portMsg := fmt.Sprintf("%v %v", CtrlClientPrefix, client.ListeningPort)
	if err := client.sendCtrl(ToNetstring(portMsg)); err != nil {
		return fmt.Errorf(errMsg, err)
	}
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
