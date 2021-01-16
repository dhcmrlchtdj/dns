package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/miekg/dns"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/dhcmrlchtdj/shunt/client"
	"github.com/dhcmrlchtdj/shunt/config"
)

type Shunt struct {
	server dns.Server
	client client.DNSClient
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg := new(config.Config)
	cfg.Load("./config.json")

	switch cfg.LogLevel {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}

	dnsMux := dns.NewServeMux()
	s := Shunt{
		server: dns.Server{
			Addr:    ":" + strconv.Itoa(cfg.Port),
			Net:     "udp",
			Handler: dnsMux,
		},
	}
	dnsMux.HandleFunc(".", s.handleRequest)
	s.client.Init(cfg.Forward)

	log.Info().Str("module", "main").Int("port", cfg.Port).Msg("Start DNS server")
	err := s.server.ListenAndServe()
	if err != nil {
		panic(err)
	}
	defer s.server.Shutdown()
}

func (s *Shunt) handleRequest(w dns.ResponseWriter, query *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(query)

	if query.Opcode == dns.OpcodeQuery {
		s.Query(m)
	}

	w.WriteMsg(m)
}

func (s *Shunt) Query(m *dns.Msg) {
	for _, q := range m.Question {
		answers := s.client.Query(q.Name, q.Qtype)
		for _, ans := range answers {
			record := fmt.Sprintf("%s %d %s %s", ans.Name, ans.TTL, dns.Type(ans.Type).String(), ans.Data)
			rr, err := dns.NewRR(record)
			if err != nil {
				panic(err)
			}
			m.Answer = append(m.Answer, rr)
		}
	}
}
