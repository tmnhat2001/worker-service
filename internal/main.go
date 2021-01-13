package main

import (
	"log"

	"github.com/tmnhat2001/worker-service/internal/api"
)

func main() {
	err := api.RunApp()
	if err != nil {
		log.Fatal(err)
	}
}
