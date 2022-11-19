package client

import (
	"context"
	"sync"

	"github.com/miekg/dns"
	"github.com/rs/zerolog"

	"github.com/dhcmrlchtdj/godns/config"
)

type DnsResolver interface {
	Resolve(ctx context.Context, question dns.Question, dnssec bool) ([]dns.RR, error)
}

var resolverCache = new(sync.Map)

func GetByUpstream(ctx context.Context, upstream *config.Upstream) DnsResolver {
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
		return createIpv4Resolver(ctx, upstream)
	}
	if upstream.Ipv6 != "" {
		return createIpv6Resolver(ctx, upstream)
	}
	if upstream.Udp != "" {
		return &Udp{server: upstream.Udp}
	}
	if upstream.Doh != "" {
		return createDohResolver(ctx, upstream)
	}

	zerolog.Ctx(ctx).Error().Str("module", "client.main").Msg("no upstream")

	return nil
}
