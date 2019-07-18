package main

import (
	"asciiclient/client"
	"fmt"
	"github.com/urfave/cli"
	"log"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "asciiclient"
	app.Usage = "./asciiclient -m <message> <nick> <destination>"

	app.Flags = []cli.Flag {
		cli.StringFlag{
			Name: "message, m",
			Value: "",
			Usage: "message string",
		},
	}

	app.Action = func(c *cli.Context) error {
		fmt.Println(c.Args())
		if len(c.String("message")) == 0 {
			log.Fatal("tried to send empty string")
		}
		asciiClient := client.New(
			c.Args().Get(0),
			c.Args().Get(1),
			1337)
		err := asciiClient.SendString(c.String("message"))
		if err != nil {
			fmt.Println(err)
		}
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
