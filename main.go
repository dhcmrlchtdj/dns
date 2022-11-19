package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"

	"github.com/dhcmrlchtdj/godns/server"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack // nolint:reassign
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	dnsServer := server.NewDnsServer()
	dnsServer.ParseArgs()

	zerolog.SetGlobalLevel(string2level(dnsServer.Config.LogLevel))

	dnsServer.SetupRouter()
	dnsServer.SetupServer()
	dnsServer.Start()
}

func string2level(s string) zerolog.Level {
	switch s {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		panic("invalid log level: " + s)
	}
}
