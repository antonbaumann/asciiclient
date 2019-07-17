package client

import (
	"fmt"
	"syscall"
)

const (
	ErrorPrefix     = "E"
	ProtocolVersion = "GRNVS V:1.0"
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

//New creates a new Client
func New(nickname string) *Model {
	return &Model{
		Nickname: nickname,
		buffer:   make([]byte, 4096),
	}
}

//Connect creates the control channel to the remote host and negotiates the data channel
func (client *Model) Connect(addr string, port int) error {
	errMsg := "connect error: %v"

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

	fmt.Println(client.token)

	if err := client.createDataSocket(); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	return nil
}
