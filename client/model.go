package client

import (
	"fmt"
	"github.com/golang/glog"
	"strings"
	"syscall"
	"time"
)

const (
	ErrorPrefix     = "E"
	ProtocolVersion = "GRNVS V:1.0"
)

type Model struct {
	Nickname       string
	RemotePort     int
	ListeningPort  int
	ctrlSocket     int
	dataSocket     int
	token          string
	sockAddrRemote syscall.Sockaddr

	buffer      []byte
}

//New creates a new Client
func New(nickname string) *Model {
	return &Model{
		Nickname:    nickname,
		buffer:      make([]byte, 4096),
	}
}

//Connect creates the control channel to the remote host and negotiates the data channel
func (client *Model) Connect(addr string, port int) error {
	errMsg := "connect error: %v"

	glog.Info("started connection")

	if err := client.dial(addr, port); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	if err := client.exchangeProtocolVersion(); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	if err := client.sendNickname(); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	if err := client.recvToken(); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	if err := client.createListeningSocket(); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	if err := client.sendListeningPortNumber(); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	return nil
}

func (client *Model) Send(message string) error {
	errMsg := "send error: %v"
	if err := client.recvProtocolConfirmation(); err != nil {
		return fmt.Errorf(errMsg, err)
	}
	return nil
}

func parseServerMessage(message, prefix string) (string, error) {
	lst := strings.SplitN(message, " ", 2)
	if len(lst) < 2 {
		return "", fmt.Errorf("server response malformed: %v", message)
	}
	if lst[0] == ErrorPrefix {
		return "", fmt.Errorf("server error: %v", message)
	}
	if lst[0] == prefix {
		return lst[1], nil
	}
	return "", fmt.Errorf("server response malformed: %v", message)
}

func (client *Model) recv(socket int, timeout time.Duration, serverPrefix string) (string, error) {
	errMsg := "receive error: %v"

	type response struct {
		length int
		err    error
	}
	ch := make(chan *response, 1)

	go func() {
		length, _, err := syscall.Recvfrom(socket, client.buffer, 0)
		ch <- &response{length, err}
	}()

	select {
	case resp := <-ch:
		if resp.err != nil {
			return "", fmt.Errorf(errMsg, resp.err)
		}
		netstring := string(client.buffer[:resp.length])
		message, err := FromNetstring(netstring)
		if err != nil {
			return message, fmt.Errorf(errMsg, err)
		}
		return parseServerMessage(message, serverPrefix)
	case <-time.After(timeout):
		return "", fmt.Errorf(errMsg, "timeout")
	}
}
