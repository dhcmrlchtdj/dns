package client

import (
	"fmt"

	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type StaticIp struct {
	addr string
}

func (ip *StaticIp) Resolve(question dns.Question) ([]dns.RR, error) {
	logger := log.With().
		Str("module", "client.static").
		Str("domain", question.Name).
		Str("record", dns.TypeToString[question.Qtype]).
		Logger()

	record := fmt.Sprintf(
		"%s %d %s %s",
		question.Name,
		60,
		dns.TypeToString[question.Qtype],
		ip.addr,
	)
	rr, err := dns.NewRR(record)
	if err != nil {
		logger.Error().Err(err).Send()
		return nil, err
	}

	logger.Debug().Msg("resolved")
	return []dns.RR{rr}, nil
}
