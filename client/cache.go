package client

import (
	"math"
	"time"

	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type dnsCached struct {
	answer  []dns.RR
	expired time.Time
}

func (c *DNSClient) cacheSet(key string, answer []dns.RR) {
	if len(answer) == 0 {
		return
	}

	minTTL := answer[0].Header().Ttl
	for _, ans := range answer {
		currTtl := ans.Header().Ttl
		if currTtl < minTTL {
			minTTL = currTtl
		}
	}

	val := dnsCached{
		answer:  answer,
		expired: time.Now().Add(time.Duration(minTTL) * time.Second),
	}
	c.cache.Store(key, &val)
}

func (c *DNSClient) cacheGet(key string) ([]dns.RR, bool) {
	val, found := c.cache.Load(key)
	if !found {
		return nil, false
	}

	cached, ok := val.(*dnsCached)
	if !ok {
		c.cache.Delete(key)
		return nil, false
	}

	elapsed := time.Until(cached.expired)
	ttl := uint32(math.Ceil(elapsed.Seconds()))
	if ttl <= 0 {
		log.Debug().Str("module", "client.cache").Str("key", key).Msg("expired")
		c.cache.Delete(key)
		return nil, false
	}

	for idx := range cached.answer {
		cached.answer[idx].Header().Ttl = ttl
	}

	return cached.answer, true
}
