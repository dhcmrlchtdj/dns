package client

import (
	"net"

	"github.com/dhcmrlchtdj/godns/config"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type Ipv4 struct {
	ip net.IP
}

func (ip *Ipv4) Resolve(question dns.Question) ([]dns.RR, error) {
	logger := log.With().
		Str("module", "client.ipv4").
		Str("domain", question.Name).
		Str("record", dns.TypeToString[question.Qtype]).
		Logger()

	rr := new(dns.A)
	rr.Hdr = dns.RR_Header{Name: question.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60}
	rr.A = ip.ip

	logger.Debug().Msg("resolved")
	return []dns.RR{rr}, nil
}

func createIpv4Resolver(upstream *config.Upstream) *Ipv4 {
	logger := log.With().
		Str("module", "client.ipv4").
		Logger()

	cacheKey := upstream.Ipv4
	if client, found := resolverCache.Load(cacheKey); found {
		logger.Trace().Msg("get resolver from cache")
		return client.(*Ipv4)
	} else {
		client := &Ipv4{ip: net.ParseIP(upstream.Ipv4)}
		resolverCache.Store(cacheKey, client)
		logger.Trace().Msg("new resolver created")
		return client
	}
}
