package server

import (
	"math"
	"time"

	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type cachedAnswer struct {
	answer  []dns.RR
	expired time.Time
}

func (s *DnsServer) cacheSet(key string, answer []dns.RR) {
	if len(answer) == 0 {
		return
	}

	minTtl := answer[0].Header().Ttl
	for _, ans := range answer {
		currTtl := ans.Header().Ttl
		if currTtl < minTtl {
			minTtl = currTtl
		}
	}

	val := cachedAnswer{
		answer:  answer,
		expired: time.Now().Add(time.Duration(minTtl) * time.Second),
	}

	log.Trace().Str("module", "server.cache").Str("key", key).Msg("added")
	s.cache.Store(key, &val)
}

func (s *DnsServer) cacheGet(key string) ([]dns.RR, bool) {
	logger := log.With().Str("module", "server.cache").Str("key", key).Logger()

	val, found := s.cache.Load(key)
	if !found {
		logger.Trace().Msg("missed")
		return nil, false
	}

	cached, ok := val.(*cachedAnswer)
	if !ok {
		s.cache.Delete(key)
		logger.Trace().Msg("missed")
		return nil, false
	}

	elapsed := time.Until(cached.expired)
	ttl := math.Ceil(elapsed.Seconds())
	if ttl <= 0 {
		s.cache.Delete(key)
		logger.Trace().Msg("expired")
		return nil, false
	}

	for idx := range cached.answer {
		cached.answer[idx].Header().Ttl = uint32(ttl)
	}

	logger.Debug().Float64("TTL", ttl).Msg("hit")
	return cached.answer, true
}
