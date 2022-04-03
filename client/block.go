package client

import (
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

func nodata(name string, qtype uint16) []dns.RR {
	sublogger := log.With().
		Str("module", "client.block").
		Str("server", "nodata").
		Str("domain", name).
		Uint16("type", qtype).
		Logger()

	sublogger.Info().Msg("query")

	return nil
}

func GetBlockNoDataClient() dnsClient {
	return nodata
}
