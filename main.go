package main

import (
	"github.com/rs/zerolog"

	"github.com/dhcmrlchtdj/godns/server"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	dnsServer := server.NewDnsServer()
	dnsServer.ParseArgs()

	zerolog.SetGlobalLevel(string2level(dnsServer.Config.LogLevel))

	dnsServer.InitRouter()
	dnsServer.InitServer()
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
