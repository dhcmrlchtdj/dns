package client

import (
	"net"

	"github.com/dhcmrlchtdj/godns/config"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type Ipv6 struct {
	ip net.IP
}

func (ip *Ipv6) Resolve(question dns.Question, dnssec bool) ([]dns.RR, error) {
	logger := log.With().
		Str("module", "client.ipv6").
		Str("domain", question.Name).
		Str("record", dns.TypeToString[question.Qtype]).
		Logger()

	rr := new(dns.AAAA)
	rr.Hdr = dns.RR_Header{Name: question.Name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 60}
	rr.AAAA = ip.ip

	logger.Debug().Msg("resolved")
	return []dns.RR{rr}, nil
}

func createIpv6Resolver(upstream *config.Upstream) *Ipv6 {
	logger := log.With().
		Str("module", "client.ipv6").
		Logger()

	cacheKey := upstream.Ipv6
	if client, found := resolverCache.Load(cacheKey); found {
		logger.Trace().Msg("get resolver from cache")
		return client.(*Ipv6)
	} else {
		client := &Ipv6{ip: net.ParseIP(upstream.Ipv6)}
		resolverCache.Store(cacheKey, client)
		logger.Trace().Msg("new resolver created")
		return client
	}
}
