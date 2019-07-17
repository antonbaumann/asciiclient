package client

import (
	"fmt"
	"syscall"
)

const (
	DataClientPrefix = "D"
	DataServerPrefix = "T"
)

//createDataSocket creates a socket for the data channel
func (client *Model) createDataSocket() error {
	fd, err := syscall.Socket(syscall.AF_INET6, syscall.SOCK_STREAM, 0)
	if err != nil {
		return err
	}

	port := bindRandomPort(fd)

	fmt.Println(port)

	client.dataSocket = fd
	return nil
}

func bindRandomPort(socket int) int {
	var inaddr6Any [16]byte
	port := getRandomUserPort()
	sa := &syscall.SockaddrInet6{
		Port: port,
		Addr: inaddr6Any,
	}
	err := syscall.Bind(socket, sa)
	if err != nil {
		fmt.Println(err)
	}
	return port
}
