package main

import (
	"log"

	"github.com/tmnhat2001/worker-service/client/wkct"
)

func main() {
	cli, err := wkct.NewCLI()
	if err != nil {
		log.Fatal(err)
	}

	cli.Run()
}
