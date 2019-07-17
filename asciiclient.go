package main

import (
	"asciiclient/client"
	"fmt"
)

func main() {
	asciiClient := client.New("h4ck3rPsch0rr")
	err := asciiClient.Connect("2a00:4700:0:9:f::c", 1337)
	if err != nil {
		fmt.Println(err)
	}

}
