package client

import (
	"github.com/dhcmrlchtdj/godns/config"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type DnsResolver interface {
	Resolve(question dns.Question) ([]dns.RR, error)
}

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
		return &StaticIp{addr: upstream.Ipv4}
	}
	if upstream.Ipv6 != "" {
		return &StaticIp{addr: upstream.Ipv6}
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
