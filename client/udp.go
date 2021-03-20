package client

import (
	"strings"
	"sync"

	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

var udpClientCache = new(sync.Map)

func GetUDPClient(udpServer string) dnsClient {
	c, found := dohClientCache.Load(udpServer)
	if found {
		return c.(dnsClient)
	}

	cc := func(name string, qtype uint16) []Answer {
		sublogger := log.With().
			Str("module", "client.udp").
			Str("server", udpServer).
			Str("domain", name).
			Uint16("type", qtype).
			Logger()

		sublogger.Debug().Msg("query")

		msg := new(dns.Msg)
		msg.SetQuestion(name, qtype)
		in, err := dns.Exchange(msg, udpServer)
		if err != nil {
			sublogger.Error().Err(err).Send()
			return nil
		}

		var ans []Answer
		for _, rr := range in.Answer {
			ans = append(ans, rr2ans(rr))
		}
		return ans
	}

	log.Debug().Str("module", "client.udp").Str("server", udpServer).Msg("create UDP server")
	udpClientCache.Store(udpServer, cc)
	return cc
}

func rr2ans(rr dns.RR) Answer {
	hd := rr.Header()
	var a Answer
	a.Name = hd.Name
	a.Type = hd.Rrtype
	a.TTL = int(hd.Ttl)
	// TODO: how to extract Data from RR
	a.Data = strings.TrimSpace(rr.String()[len(hd.String()):])
	return a
}
