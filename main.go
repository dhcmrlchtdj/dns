package main

import (
	"github.com/dhcmrlchtdj/godns/server"
	"github.com/rs/zerolog/log"
)

func main() {
	dnsServer := server.NewDnsServer()
	dnsServer.ParseArgs()
	dnsServer.InitDnsRouter()
	dnsServer.InitDnsServer()
	defer dnsServer.Shutdown()
	if err := dnsServer.ListenAndServe(); err != nil {
		log.Error().Err(err).Send()
		panic(err)
	}
}
