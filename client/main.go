package client

import (
	"sync"

	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"

	"github.com/dhcmrlchtdj/godns/config"
)

type DnsResolver interface {
	Resolve(question dns.Question, dnssec bool) ([]dns.RR, error)
}

var resolverCache = new(sync.Map)

func GetByUpstream(upstream *config.Upstream) DnsResolver {
	if upstream == nil {
		return nil
	}

	if upstream.Block == "nodata" {
		return &BlockByNodata{}
	}
	if upstream.Block == "nxdomain" {
		return &BlockByNxdomain{}
	}
	if upstream.Ipv4 != "" {
		return createIpv4Resolver(upstream)
	}
	if upstream.Ipv6 != "" {
		return createIpv6Resolver(upstream)
	}
	if upstream.Udp != "" {
		return &Udp{server: upstream.Udp}
	}
	if upstream.Doh != "" {
		return createDohResolver(upstream)
	}

	log.Error().Str("module", "client.main").Msg("no upstream")

	return nil
}
