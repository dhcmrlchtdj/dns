package main

import (
	"github.com/dhcmrlchtdj/godns/server"
	"github.com/rs/zerolog/log"
)

func main() {
	server := server.NewDnsServer()
	server.ParseArgs()
	server.InitDnsRouter()
	server.InitDnsServer()
	if err := server.ListenAndServe(); err != nil {
		log.Error().Err(err).Send()
		panic(err)
	}
	defer server.Shutdown()
}
