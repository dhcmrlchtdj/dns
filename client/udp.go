package client

import (
	"sync"

	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

var udpClientCache = new(sync.Map)

func GetUDPClient(udpServer string) dnsClient {
	c, found := udpClientCache.Load(udpServer)
	if found {
		return c.(dnsClient)
	}

	cc := func(name string, qtype uint16) []dns.RR {
		sublogger := log.With().
			Str("module", "client.udp").
			Str("server", udpServer).
			Str("domain", name).
			Uint16("type", qtype).
			Logger()

		sublogger.Info().Msg("query")

		msg := new(dns.Msg)
		msg.SetQuestion(name, qtype)
		in, err := dns.Exchange(msg, udpServer)
		if err != nil {
			sublogger.Error().Err(err).Send()
			return nil
		}
		return in.Answer
	}

	log.Debug().Str("module", "client.udp").Str("server", udpServer).Msg("create UDP server")
	udpClientCache.Store(udpServer, cc)
	return cc
}
