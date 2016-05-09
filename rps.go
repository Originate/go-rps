package main

import (
	. "github.com/Originate/go_rps/server"
	"log"
)

func main() {
	server := GoRpsServer{}
	serverTCPAddr, err := server.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Server running on: %s\n", serverTCPAddr.String())
	select {}
}
