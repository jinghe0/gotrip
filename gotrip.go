package main

import (
	"flag"
	"github.com/vtfr/gotrip/manager"
)

func main() {
	var config struct {
		JackName string
		IsClient bool
		Address  string
	}

	flag.BoolVar(&config.IsClient, "client-mode", false, "is client mode")
	flag.StringVar(&config.JackName, "jack", "GoTrip", "jack client name")
	flag.StringVar(&config.Address, "address", ":4466", "endereço para binding")
	flag.Parse()

	manager.NewManager(config.JackName, config.Address, config.IsClient).
		Start()
}
