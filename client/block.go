package client

import (
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type BlockByNodata struct{}

func (*BlockByNodata) Resolve(question dns.Question) ([]dns.RR, error) {
	log.Debug().
		Str("module", "client.block.nodata").
		Str("domain", question.Name).
		Str("record", dns.TypeToString[question.Qtype]).
		Msg("resolved")
	return nil, nil
}

type BlockByNxdomain struct{}

func (*BlockByNxdomain) Resolve(question dns.Question) ([]dns.RR, error) {
	log.Debug().
		Str("module", "client.block.nxdomain").
		Str("domain", question.Name).
		Str("record", dns.TypeToString[question.Qtype]).
		Msg("resolved")
	return nil, &ErrDnsResponse{Rcode: dns.RcodeNameError}
}
