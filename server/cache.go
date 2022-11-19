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
	logger := zerolog.Ctx(s.ctx).
		With().
		Str("module", "server.cache.cleanup").
		Logger()

	ticker := time.NewTicker(time.Minute * 10)
	for {
		select {
		case <-ticker.C:
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
		case <-s.ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (s *DnsServer) cacheSet(ctx context.Context, key string, answer []dns.RR) {
	if len(answer) == 0 {
		return
	}

	// set ttl to the minimum ttl among all answers
	ttl := answer[0].Header().Ttl
	for _, ans := range answer {
		currTtl := ans.Header().Ttl
		if currTtl < ttl {
			ttl = currTtl
		}
	}
	// limit the max ttl to 1 hour
	maxTtl := uint32(60 * 60)
	if ttl > maxTtl {
		ttl = maxTtl
	}

	val := cachedAnswer{
		answer:  answer,
		expired: time.Now().Add(time.Duration(ttl) * time.Second),
	}

	zerolog.Ctx(ctx).
		Trace().
		Str("module", "server.cache").
		Str("key", key).
		Uint32("TTL", ttl).
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
