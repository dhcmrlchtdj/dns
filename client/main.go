package client

import (
	"math"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/miekg/dns"

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
			println("unsupported scheme") // TODO
			continue
		}

		for _, domain := range forward.Domain {
			c.router.add(dns.Fqdn(domain), cli)
		}
	}
}

///

func (c *DNSClient) Query(name string, qtype uint16) []Answer {
	println("query..........", name)
	cacheKey := name + "|" + strconv.Itoa(int(qtype))

	// from cache
	cached, found := c.cacheGet(cacheKey)
	if found {
		return cached
	}

	// by config
	println("query..........", name)
	cli := c.router.route(dns.Fqdn(name))
	if cli == nil {
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
