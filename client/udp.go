package client

import (
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type Udp struct {
	server string
}

func (u *Udp) Resolve(question dns.Question) ([]dns.RR, error) {
	logger := log.With().
		Str("module", "client.udp").
		Str("domain", question.Name).
		Str("record", dns.TypeToString[question.Qtype]).
		Logger()

	msg := new(dns.Msg)
	msg.SetQuestion(question.Name, question.Qtype)
	in, err := dns.Exchange(msg, u.server)
	if err != nil {
		logger.Error().Err(err).Send()
		return nil, err
	}

	logger.Debug().Msg("resolved")
	return in.Answer, nil
}
