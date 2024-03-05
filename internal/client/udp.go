package client

import (
	"context"

	"github.com/miekg/dns"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type Udp struct {
	server string
}

func (u *Udp) Resolve(ctx context.Context, question dns.Question, dnssec bool) ([]dns.RR, error) {
	logger := zerolog.Ctx(ctx).
		With().
		Str("module", "client.udp").
		Str("domain", question.Name).
		Str("record", dns.TypeToString[question.Qtype]).
		Logger()

	msg := new(dns.Msg)
	msg.SetQuestion(question.Name, question.Qtype)
	if dnssec {
		msg.SetEdns0(4096, true)
	}
	in, err := dns.ExchangeContext(ctx, msg, u.server)
	if err != nil {
		err = errors.WithStack(err)
		logger.Error().Stack().Err(err).Send()
		return nil, err
	}

	logger.Debug().Msg("resolved")
	return in.Answer, nil
}
