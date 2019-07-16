package main

import (
	"asciiclient/client"
	"fmt"
)

func main() {
	asciiClient := client.New("rüdiger")
	err := asciiClient.Connect("2a00:4700:0:9:f::c", 1337)
	if err != nil {
		fmt.Println(err)
	}
}
