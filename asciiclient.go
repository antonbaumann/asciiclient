package main

import (
	"asciiclient/client"
	"fmt"
)

func main() {
	asciiClient := client.New("h4ck3rPsch0rr", "2a00:4700:0:9:f::c", 1337)
	err := asciiClient.SendString("m")
	if err != nil {
		fmt.Println(err)
	}
}
