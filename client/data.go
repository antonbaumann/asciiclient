package client

import (
	"fmt"
	"github.com/golang/glog"
	"golang.org/x/sys/unix"
	"strconv"
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
	fd, err := unix.Socket(unix.AF_INET6, unix.SOCK_STREAM, 0)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	port, err := bindRandomPort(fd)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	if err := unix.Listen(fd, 1); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	client.listeningPort = port
	client.dataSocket = fd

	if err := client.sendListeningPortNumber(); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	return nil
}

func (client *Model) sendListeningPortNumber() error {
	errMsg := "[data] send data port error: %v"
	if err := client.sendCtrl(strconv.Itoa(client.listeningPort)); err != nil {
		return fmt.Errorf(errMsg, err)
	}
	glog.Infof("sent listening port [%v] to server", client.listeningPort)
	glog.Infof("listening now on port %v", client.listeningPort)
	return nil
}

func bindRandomPort(socket int) (int, error) {
	errMsg := "[data] bind port to data socket error: %v"
	var inaddr6Any [16]byte
	var err error
	nrRetries := 5
	for i := 0; i < nrRetries; i++ {
		port := GetRandomUserPort()
		sa := &unix.SockaddrInet6{
			Port: port,
			Addr: inaddr6Any,
		}
		if err = unix.Bind(socket, sa); err == nil {
			return port, nil
		}
	}
	return -1, fmt.Errorf(errMsg, err)
}

func (client *Model) awaitServerConnection() (unix.Sockaddr, error) {
	errMsg := "[data] await server connection: %v"
	nfd, sa, err := unix.Accept(client.dataSocket)
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}
	client.dataSocket = nfd
	return sa, nil
}

func (client *Model) recvProtocolConfirmation(sa unix.Sockaddr) error {
	errMsg := "receive protocol confirmation error: %v"
	msg, err := client.recvData()
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	if msg != ProtocolVersion {
		return fmt.Errorf(errMsg, "protocols do not match")
	}
	glog.Info("received protocol confirmation")
	return nil
}

func (client *Model) recvData() (string, error) {
	errMsg := "[data] %v"
	msg, err := client.recv(client.dataSocket, DataReceiveTimeout, DataServerPrefix)
	if err != nil {
		_ = client.sendDataError(err.Error())
		return msg, fmt.Errorf(errMsg, err)
	}
	return msg, nil
}

//sendCtrl is a helper method for sending strings over the control channel
func (client *Model) sendData(message string) error {
	errMsg := "[data] send error: %v"
	netstring := ToNetstring(fmt.Sprintf("%v %v", DataClientPrefix, message))
	err := unix.Sendto(client.dataSocket, []byte(netstring), 0, client.sockAddrRemote)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	return nil
}

func (client *Model) sendDataError(message string) error {
	errMsg := "[data] send error: %v"
	netstring := ToNetstring(fmt.Sprintf("%v %v", ErrorPrefix, message))
	err := unix.Sendto(client.dataSocket, []byte(netstring), 0, client.sockAddrRemote)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	return nil
}

func (client *Model) validateToken() error {
	errMsg := "[data] error receiving token: %v"
	token, err := client.recvData()
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	if client.token != token {
		return fmt.Errorf(errMsg, "server sent wrong token")
	}
	glog.Info("server sent valid token")
	return nil
}
