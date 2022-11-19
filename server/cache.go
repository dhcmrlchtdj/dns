package server

import (
	"context"
	"math"
	"time"

	"github.com/miekg/dns"
	"github.com/rs/zerolog"
)

type cachedAnswer struct {
	answer  []dns.RR
	expired time.Time
}

func (s *DnsServer) cleanupExpiredCache() {
	ticker := time.NewTicker(time.Minute)

	go func() {
		logger := zerolog.Ctx(s.ctx).
			With().
			Str("module", "server.cache.cleanup").
			Logger()

		for range ticker.C {
			logger.Trace().Msg("cleaning")
			s.cache.Range(func(key any, val any) bool {
				cached, ok := val.(*cachedAnswer)
				if !ok {
					s.cache.Delete(key)
					logger.Trace().Str("key", key.(string)).Msg("invalid")
					return true
				}

				sec := math.Ceil(time.Until(cached.expired).Seconds())
				if sec <= 0 {
					s.cache.Delete(key)
					logger.Trace().Str("key", key.(string)).Msg("expired")
					return true
				}

				return true
			})
			logger.Trace().Msg("cleaned")
		}
	}()
}

func (s *DnsServer) cacheSet(ctx context.Context,key string, answer []dns.RR) {
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

	zerolog.Ctx(ctx).
		Trace().
		Str("module", "server.cache").
		Str("key", key).
		Uint32("TTL", minTtl).
		Msg("added")
	s.cache.Store(key, &val)
}

func (s *DnsServer) cacheGet(ctx context.Context, key string) []dns.RR {
	logger := zerolog.Ctx(ctx).
		With().
		Str("module", "server.cache").
		Str("key", key).
		Logger()

	val, found := s.cache.Load(key)
	if !found {
		logger.Trace().Msg("missed")
		return nil
	}

	cached, ok := val.(*cachedAnswer)
	if !ok {
		s.cache.Delete(key)
		logger.Trace().Msg("invalid")
		return nil
	}

	sec := math.Ceil(time.Until(cached.expired).Seconds())
	if sec <= 0 {
		s.cache.Delete(key)
		logger.Trace().Msg("expired")
		return nil
	}
	ttl := uint32(sec)

	for idx := range cached.answer {
		cached.answer[idx].Header().Ttl = ttl
	}

	logger.Debug().Uint32("TTL", ttl).Msg("hit")
	return cached.answer
}
