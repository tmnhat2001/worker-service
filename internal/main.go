package main

import (
	"log"

	"github.com/tmnhat2001/worker-service/internal/api"
)

const certPath = "certs/server.crt"
const keyPath = "certs/server.key"

func main() {
	config := api.ServerConfig{
		Port:         8080,
		CertFilePath: certPath,
		KeyFilePath:  keyPath,
	}
	server, err := api.NewServer(config)
	if err != nil {
		log.Fatal(err)
	}

	err = server.Run()
	if err != nil {
		log.Fatal(err)
	}
}
