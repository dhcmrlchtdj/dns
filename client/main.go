package client

import (
	"math"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"

	"github.com/dhcmrlchtdj/shunt/config"
)

///

type dnsClient func(string, uint16) []Answer

type DNSClient struct {
	cache  sync.Map // MAP("domain|type") => dnsCached
	router dnsRouter
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
		case "udp":
			cli = GetUDPClient(parsed.Host)
		case "doh":
			parsed.Scheme = "https"
			cli = GetDoHClient(parsed.String())
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
	cacheKey := name + "|" + strconv.Itoa(int(qtype))
	log.Info().Str("module", "client").Str("domain", name).Uint16("type", qtype).Msg("query")

	// from cache
	cached, found := c.cacheGet(cacheKey)
	if found {
		log.Debug().Str("module", "client").Str("domain", name).Uint16("type", qtype).Msg("cache hit")
		return cached
	}

	// by config
	cli := c.router.route(dns.Fqdn(name))
	if cli == nil {
		log.Debug().Str("module", "client").Str("domain", name).Uint16("type", qtype).Msg("not found")
		return nil
	}
	ans := cli(name, qtype)
	c.cacheSet(cacheKey, ans)
	return ans
}

///

type dnsCached struct {
	answer  []Answer
	expired time.Time
}

func (c *DNSClient) cacheSet(key string, answer []Answer) {
	if len(answer) == 0 {
		return
	}

	ttl := time.Duration(answer[0].TTL) * time.Second

	val := dnsCached{
		answer:  answer,
		expired: time.Now().Add(ttl),
	}
	c.cache.Store(key, &val)
}

func (c *DNSClient) cacheGet(key string) ([]Answer, bool) {
	val, found := c.cache.Load(key)
	if !found {
		return nil, false
	}

	cached, ok := val.(*dnsCached)
	if !ok {
		c.cache.Delete(key)
		return nil, false
	}

	elapsed := cached.expired.Sub(time.Now())
	ttl := int(math.Ceil(elapsed.Seconds()))
	if ttl <= 0 {
		log.Trace().Str("module", "client.cache").Str("key", key).Msg("expired")
		c.cache.Delete(key)
		return nil, false
	}

	for idx := range cached.answer {
		cached.answer[idx].TTL = ttl
	}

	return cached.answer, true
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
