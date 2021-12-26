package client

import (
	"net/url"
	"strconv"
	"sync"

	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"

	"github.com/dhcmrlchtdj/godns/config"
)

///

type dnsClient func(string, uint16) []Answer

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

func (c *DNSClient) Query(name string, qtype uint16) []Answer {
	log.Debug().Str("module", "client").Str("domain", name).Uint16("type", qtype).Msg("query")

	name = dns.Fqdn(name)

	// from staticIp
	if qtype == dns.TypeA {
		staticIp, found := c.staticIpV4[name]
		if found {
			log.Info().Str("module", "client").Str("domain", name).Uint16("type", qtype).Msg("staticIpV4 hit")
			return []Answer{{Name: name, Type: qtype, TTL: 60, Data: staticIp}}
		}
	} else if qtype == dns.TypeAAAA {
		staticIp, found := c.staticIpV6[name]
		if found {
			log.Info().Str("module", "client").Str("domain", name).Uint16("type", qtype).Msg("staticIpV6 hit")
			return []Answer{{Name: name, Type: qtype, TTL: 60, Data: staticIp}}
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

///

type Answer struct {
	// The record owner.
	Name string `json:"name"`
	// The type of DNS record.
	Type uint16 `json:"type"`
	// The number of seconds the answer can be stored in cache before it is considered stale.
	TTL int `json:"TTL"`
	// The value of the DNS record for the given name and type.
	Data string `json:"data"`
}
