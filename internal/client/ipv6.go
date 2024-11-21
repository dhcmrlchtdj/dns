package client

import (
	"context"
	"net"

	"github.com/miekg/dns"
	"github.com/rs/zerolog"
)

type Ipv6 struct {
	ip net.IP
}

func (ip *Ipv6) Resolve(ctx context.Context, question dns.Question, dnssec bool) ([]dns.RR, error) {
	logger := zerolog.Ctx(ctx).
		With().
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

func createIpv6Resolver(ctx context.Context, ip string) DnsResolver {
	logger := zerolog.Ctx(ctx).
		With().
		Str("module", "client.ipv6").
		Logger()

	if client, found := resolverCache.Get(ip); found {
		logger.Trace().Msg("get resolver from cache")
		return client
	} else {
		client := &Ipv6{ip: net.ParseIP(ip)}
		resolverCache.Set(ip, client)
		logger.Trace().Msg("new resolver created")
		return client
	}
}
