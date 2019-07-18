package client

import (
	"fmt"
	"github.com/golang/glog"
	"golang.org/x/sys/unix"
	"strings"
	"time"
)

const (
	ErrorPrefix     = "E"
	ProtocolVersion = "GRNVS V:1.0"
)

type Model struct {
	Nickname       string
	remotePort     int
	remoteIP       string
	listeningPort  int
	ctrlSocket     int
	dataSocket     int
	token          string
	lastMessageLen int
	sockAddrRemote unix.Sockaddr

	buffer []byte
}

//New creates a new Client
func New(nickname string, remoteIP string, remotePort int) *Model {
	return &Model{
		Nickname:   nickname,
		buffer:     make([]byte, 4096),
		remotePort: remotePort,
		remoteIP:   remoteIP,
	}
}

func handle(err error) error {
	fmt.Printf("Error: %v\n", err)
	return err
}

func (client *Model) SendString(message string) error {
	if err := client.Connect(); err != nil {
		return handle(err)
	}
	if err := client.Send(message); err != nil {
		return handle(err)
	}
	if err := client.Disconnect(); err != nil {
		return handle(err)
	}
	return nil
}

//Connect creates the control channel to the remote host and negotiates the data channel
func (client *Model) Connect() error {
	errMsg := "connect error: %v"

	glog.Info("started connection")

	if err := client.dial(); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	if err := client.exchangeProtocolVersion(); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	// send client nickname to server
	if err := client.sendCtrl(client.Nickname); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	// receive token from server
	token, err := client.recvCtrl()
	if err != nil {
		info := "[control] error receiving token: %v"
		return fmt.Errorf(errMsg, fmt.Errorf(info, err))
	}
	client.token = token

	if err := client.createListeningSocket(); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	return nil
}

func (client *Model) Send(message string) error {
	errMsg := "send error: %v"
	sa, err := client.awaitServerConnection()
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	if err := client.recvProtocolConfirmation(sa); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	if err := client.sendData(client.Nickname); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	if err := client.validateToken(); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	// transfer message
	if err := client.sendData(message); err != nil {
		return fmt.Errorf(errMsg, err)
	}
	client.lastMessageLen = len(message)
	glog.Info("message successfully transferred")

	// receive new token
	token, err := client.recvData()
	if err != nil {
		info := "[data] error receiving token: %v"
		return fmt.Errorf(errMsg, fmt.Errorf(info, err))
	}
	client.token = token
	glog.Info("received new token")

	// close data socket
	if err := unix.Close(client.dataSocket); err != nil {
		info := "[data] error closing data channel: %v"
		return fmt.Errorf(errMsg, fmt.Errorf(info, err))
	}

	return nil
}

func (client *Model) Disconnect() error {
	errMsg := "disconnect error: %v"

	if err := client.validateStringLength(); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	if err := client.sendCtrl(client.token); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	msg, err := client.recvCtrl()
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	if msg != "ACK" {
		err := fmt.Errorf("server sent [%v] should be [%v]", msg, "ACK")
		return fmt.Errorf(errMsg, err)
	}
	glog.Infof("server accepted message transfer")

	if err := unix.Close(client.ctrlSocket); err != nil {
		info := "[ctrl] error closing control channel: %v"
		return fmt.Errorf(errMsg, fmt.Errorf(info, err))
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
		length, _, err := unix.Recvfrom(socket, client.buffer, 0)
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
