package main

import (
	"flag"
	"github.com/vtfr/gotrip/manager"
	"github.com/vtfr/gotrip/audio"
	"log"
)

func main() {
	var config struct {
		JackName string
		IsClient bool
		Address  string
		NumChan int
	}

	flag.BoolVar(&config.IsClient, "client-mode", false, "is client mode")
	flag.StringVar(&config.JackName, "jack", "GoTrip", "jack client name")
	flag.IntVar(&config.NumChan, "chan", 2, "numero canais")
	flag.StringVar(&config.Address, "address", ":4466", "endereço para binding")
	flag.Parse()

	err := audio.Initialize(config.JackName)
	if err != nil {
		log.Fatalln(err)
	}

	if config.IsClient {
		s, err := manager.NewSessionAsMember(config.JackName, config.Address)
		if err != nil {
			log.Fatalln(err)
		}

		s.Run()
	} else {
		s, err := manager.NewSessionAsMaster(config.JackName, config.Address,
			config.NumChan, 0)
		if err != nil {
			log.Fatalln(err)
		}

		s.Run()
	}
}
