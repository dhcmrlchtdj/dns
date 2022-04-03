package client

import (
	"fmt"
	"net/url"
	"strconv"
	"sync"

	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"

	"github.com/dhcmrlchtdj/godns/config"
)

///

type dnsClient func(string, uint16) []dns.RR

type DNSClient struct {
	cache      sync.Map // MAP("domain|type") => dnsCached
	router     dnsRouter
	staticIpV4 map[string]string
	staticIpV6 map[string]string
}

///

func (c *DNSClient) Init(forwards []config.Server) {
	for _, forward := range forwards {

		parsed, err := url.Parse(forward.DNS)
		if err != nil {
			log.Error().Str("module", "client").Str("dns", forward.DNS).Msg("invalid config")
			panic(err)
		}
		var cli dnsClient
		switch parsed.Scheme {
		case "ipv4":
			if c.staticIpV4 == nil {
				c.staticIpV4 = make(map[string]string)
			}
			for _, domain := range forward.Domain {
				c.staticIpV4[dns.Fqdn(domain)] = parsed.Host
			}
			continue
		case "ipv6":
			if c.staticIpV6 == nil {
				c.staticIpV6 = make(map[string]string)
			}
			for _, domain := range forward.Domain {
				c.staticIpV6[dns.Fqdn(domain)] = parsed.Host
			}
			continue
		case "udp":
			cli = GetUDPClient(parsed.Host)
		case "doh":
			parsed.Scheme = "https"
			cli = GetDoHClient(parsed.String(), forward.HttpsProxy)
		case "tcp", "dot":
			log.Error().Str("module", "client").Str("dns", forward.DNS).Msg("WIP")
			continue
		case "block":
			if parsed.Hostname() == "nodata" {
				cli = GetBlockNoDataClient()
			} else {
				log.Error().Str("module", "client").Str("dns", forward.DNS).Msg("WIP")
				continue
			}
		default:
			log.Error().Str("module", "client").Str("dns", forward.DNS).Msg("unsupported scheme")
			continue
		}

		for _, domain := range forward.Domain {
			c.router.add(dns.Fqdn(domain), cli)
		}
	}
}

///

func (c *DNSClient) Query(name string, qtype uint16) []dns.RR {
	log.Debug().Str("module", "client").Str("domain", name).Uint16("type", qtype).Msg("query")

	name = dns.Fqdn(name)

	// from staticIp
	if qtype == dns.TypeA {
		staticIp, found := c.staticIpV4[name]
		if found {
			log.Info().Str("module", "client").Str("domain", name).Uint16("type", qtype).Msg("staticIpV4 hit")
			record := fmt.Sprintf("%s %d %s %s", name, 3600, dns.Type(qtype).String(), staticIp)
			rr, err := dns.NewRR(record)
			if err != nil {
				log.Error().Str("module", "client").Str("domain", name).Uint16("type", qtype).Err(err).Send()
			}
			return []dns.RR{rr}
		}
	} else if qtype == dns.TypeAAAA {
		staticIp, found := c.staticIpV6[name]
		if found {
			log.Info().Str("module", "client").Str("domain", name).Uint16("type", qtype).Msg("staticIpV6 hit")
			record := fmt.Sprintf("%s %d %s %s", name, 3600, dns.Type(qtype).String(), staticIp)
			rr, err := dns.NewRR(record)
			if err != nil {
				log.Error().Str("module", "client").Str("domain", name).Uint16("type", qtype).Err(err).Send()
			}
			return []dns.RR{rr}
		}
	}

	cacheKey := name + "|" + strconv.Itoa(int(qtype))

	// from cache
	cached, found := c.cacheGet(cacheKey)
	if found {
		log.Debug().Str("module", "client").Str("domain", name).Uint16("type", qtype).Msg("cache hit")
		return cached
	}

	// by config
	cli := c.router.route(name)
	if cli != nil {
		ans := cli(name, qtype)
		c.cacheSet(cacheKey, ans)
		return ans
	}

	// not found
	log.Info().Str("module", "client").Str("domain", name).Uint16("type", qtype).Msg("not found")
	return nil
}
