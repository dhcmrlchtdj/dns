package client

import (
	"strconv"
	"sync"
	"time"
)

///

type DNSClient struct {
	cache sync.Map // MAP(domain+type) => dnsCached
}

func (c *DNSClient) Query(name string, qtype uint16) []Answer {
	cacheKey := name + "|" + strconv.Itoa(int(qtype))

	// from cache
	cached, found := c.cacheGet(cacheKey)
	if found {
		return cached
	}

	// by config
	resp := c.queryByConfig(name, qtype)
	c.cacheSet(cacheKey, resp)
	return resp
}

func (c *DNSClient) queryByConfig(name string, qtype uint16) []Answer {
	// TODO
	return nil
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

	ttl := time.Duration(answer[0].TTL) * time.Nanosecond

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

	if time.Now().After(cached.expired) {
		c.cache.Delete(key)
		return nil, false
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
type dnsClient func(string, uint16) []Answer
