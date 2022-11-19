package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"

	"github.com/dhcmrlchtdj/godns/server"
)

func main() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack // nolint:reassign
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	dnsServer := server.NewDnsServer()
	dnsServer.SetupContext()
	dnsServer.ParseArgs()

	level, err := zerolog.ParseLevel(dnsServer.Config.LogLevel)
	if err != nil {
		panic(err)
	}
	zerolog.SetGlobalLevel(level)

	dnsServer.SetupRouter()
	dnsServer.SetupServer()
	dnsServer.Start()
}
