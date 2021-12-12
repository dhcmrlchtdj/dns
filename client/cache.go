package client

import (
	"math"
	"time"

	"github.com/rs/zerolog/log"
)

type dnsCached struct {
	answer  []Answer
	expired time.Time
}

func (c *DNSClient) cacheSet(key string, answer []Answer) {
	if len(answer) == 0 {
		return
	}

	minTTL := answer[0].TTL
	for _, ans := range answer {
		if ans.TTL < minTTL {
			minTTL = ans.TTL
		}
	}

	val := dnsCached{
		answer:  answer,
		expired: time.Now().Add(time.Duration(minTTL) * time.Second),
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
		log.Debug().Str("module", "client.cache").Str("key", key).Msg("expired")
		c.cache.Delete(key)
		return nil, false
	}

	for idx := range cached.answer {
		cached.answer[idx].TTL = ttl
	}

	return cached.answer, true
}
