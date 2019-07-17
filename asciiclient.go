package main

import (
	"asciiclient/client"
	"fmt"
)

func main() {
	asciiClient := client.New("h4ck3rPsch0rr")
	if err := asciiClient.Connect("2a00:4700:0:9:f::c", 1337); err != nil {
		fmt.Println(err)
		return
	}
	if err := asciiClient.Send("hello"); err != nil {
		fmt.Println(err)
		return
	}
}
