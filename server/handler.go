package server

import (
	"errors"

	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"

	"github.com/dhcmrlchtdj/godns/client"
)

func (s *DnsServer) handleRequest(w dns.ResponseWriter, request *dns.Msg) {
	logger := log.With().
		Str("module", "server.handler").
		Uint16("request_id", request.Id).
		Logger()

	reply := new(dns.Msg)
	reply.SetReply(request)
	if edns := request.IsEdns0(); edns != nil {
		reply.SetEdns0(4096, true)
	}

	logger.Trace().
		Str("opcode", dns.OpcodeToString[request.Opcode]).
		Msg("receive request")
	if request.Opcode == dns.OpcodeQuery {
		s.Query(reply)
	} else {
		reply.Rcode = dns.RcodeNotImplemented
	}

	err := w.WriteMsg(reply)
	if err != nil {
		logger.Error().Err(err).Msg("failed to write reply")
	}
}

func (s *DnsServer) Query(reply *dns.Msg) {
	logger := log.With().
		Str("module", "server.handler").
		Uint16("request_id", reply.Id).
		Logger()

	if len(reply.Question) != 1 {
		logger.Debug().
			Int("question", len(reply.Question)).
			Msg("format error")
		reply.Rcode = dns.RcodeFormatError
		return
	}

	question := reply.Question[0]
	logger.Info().
		Str("name", question.Name).
		Str("record", dns.TypeToString[question.Qtype]).
		Bool("dnssec", reply.IsEdns0() != nil).
		Msg("query")

	// from cache
	cacheKey := question.String()
	ans := s.cacheGet(cacheKey)
	if ans != nil {
		logger.Trace().Msg("from cache")
		reply.Answer = ans
		return
	}

	upstream := s.router.search(question.Name, question.Qtype)

	// no upstream
	if upstream == nil {
		logger.Trace().Msg("no upstream")
		reply.Rcode = dns.RcodeNotImplemented
		return
	}
	// no resolver
	resolver := client.GetByUpstream(upstream)
	if resolver == nil {
		logger.Error().Msg("no resolver")
		reply.Rcode = dns.RcodeNotImplemented
		return
	}

	// from upstream
	ans, err := resolver.Resolve(question, reply.IsEdns0() != nil)
	if err == nil {
		reply.Answer = ans
		s.cacheSet(cacheKey, ans)
		logger.Trace().Msg("resolved")
	} else {
		var errRcode *client.ErrDnsResponse
		if errors.As(err, &errRcode) {
			reply.Rcode = errRcode.Rcode
			logger.Debug().
				Str("Rcode", dns.RcodeToString[reply.Rcode]).
				Msg("resolved")
		} else {
			reply.Rcode = dns.RcodeServerFailure
			logger.Error().Err(err).Msg("unknown error")
		}
	}
}
