package server

import (
	"context"
	"math"
	"time"

	"github.com/miekg/dns"
	"github.com/rs/zerolog"

	"github.com/dhcmrlchtdj/godns/util"
)

type cachedAnswer struct {
	answer  []dns.RR
	expired time.Time
}
type deferredAnswer = util.Deferred[cachedAnswer, int]

///

func (s *DnsServer) cacheGet(ctx context.Context, key string) ([]dns.RR, *int) {
	logger := zerolog.Ctx(ctx).
		With().
		Str("module", "server.cache.get").
		Str("key", key).
		Logger()

	val, found := s.cache.Load(key)
	if !found {
		logger.Trace().Msg("missed")
		return nil, nil
	}

	deferredAnswer, ok := val.(*deferredAnswer)
	if !ok {
		s.cache.Delete(key)
		logger.Trace().Msg("invalid")
		return nil, nil
	}

	cached, rcode := deferredAnswer.Wait()
	if rcode != nil {
		return nil, rcode
	}

	sec := math.Ceil(time.Until(cached.expired).Seconds())
	if sec <= 0 {
		s.cache.Delete(key)
		logger.Trace().Msg("expired")
		return nil, nil
	}
	ttl := uint32(sec)

	for idx := range cached.answer {
		cached.answer[idx].Header().Ttl = ttl
	}

	logger.Debug().Uint32("TTL", ttl).Msg("hit")

	return cached.answer, nil
}

func (s *DnsServer) cacheSet(ctx context.Context, key string, deferredAnswer *deferredAnswer) {
	logger := zerolog.Ctx(ctx).
		With().
		Str("module", "server.cache.set").
		Str("key", key).
		Logger()

	s.cache.Store(key, deferredAnswer)

	logger.Trace().Msg("added")
}

func (s *DnsServer) cacheResolve(ctx context.Context, key string, answer []dns.RR) {
	logger := zerolog.Ctx(ctx).
		With().
		Str("module", "server.cache.resolve").
		Str("key", key).
		Logger()

	val, found := s.cache.Load(key)
	if !found {
		logger.Trace().Msg("missed")
		return
	}

	deferredAnswer, ok := val.(*deferredAnswer)
	if !ok {
		s.cache.Delete(key)
		logger.Trace().Msg("invalid")
		return
	}

	// limit the max ttl to 1 hour
	maxTtl := uint32(60 * 60)
	ttl := maxTtl
	// set ttl to the minimum ttl among all answers
	for _, ans := range answer {
		currTtl := ans.Header().Ttl
		if currTtl < ttl {
			ttl = currTtl
		}
	}

	ans := cachedAnswer{
		answer:  answer,
		expired: time.Now().Add(time.Duration(ttl) * time.Second),
	}
	deferredAnswer.Resolve(ans)

	logger.Trace().
		Uint32("TTL", ttl).
		Msg("resolved")
}

func (s *DnsServer) cacheReject(ctx context.Context, key string, rcode int) {
	logger := zerolog.Ctx(ctx).
		With().
		Str("module", "server.cache.reject").
		Str("key", key).
		Logger()

	val, found := s.cache.Load(key)
	if !found {
		logger.Trace().Msg("missed")
		return
	}

	deferredAnswer, ok := val.(*deferredAnswer)
	if !ok {
		s.cache.Delete(key)
		logger.Trace().Msg("invalid")
		return
	}

	deferredAnswer.Reject(rcode)

	logger.Trace().
		Int("rcode", rcode).
		Msg("rejected")
}

///

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
				deferredAnswer, ok := val.(*deferredAnswer)
				if !ok {
					s.cache.Delete(key)
					logger.Trace().Msg("invalid")
					return true
				}

				cached, rcode := deferredAnswer.Wait()
				if rcode != nil {
					return true
				}

				sec := math.Ceil(time.Until(cached.expired).Seconds())
				if sec <= 0 {
					s.cache.Delete(key)
					logger.Trace().Msg("expired")
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
