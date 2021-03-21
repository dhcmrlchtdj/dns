package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/miekg/dns"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/dhcmrlchtdj/dns/client"
	"github.com/dhcmrlchtdj/dns/config"
)

type Dns struct {
	server dns.Server
	client client.DNSClient
}

func main() {
	cfg := initConfig()

	dnsMux := dns.NewServeMux()
	s := Dns{
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

///

func (s *Dns) handleRequest(w dns.ResponseWriter, query *dns.Msg) {
	log.Debug().Str("module", "main").Msg("handle request")

	m := new(dns.Msg)
	m.SetReply(query)

	if query.Opcode == dns.OpcodeQuery {
		s.Query(m)
	}

	err := w.WriteMsg(m)
	if err != nil {
		log.Debug().Str("module", "main").Err(err).Msg("handle request write")
	}
}

func (s *Dns) Query(m *dns.Msg) {
	log.Debug().Str("module", "main").Msg("query")

	for _, q := range m.Question {
		answers := s.client.Query(q.Name, q.Qtype)
		for _, ans := range answers {
			record := fmt.Sprintf("%s %d %s %s", ans.Name, ans.TTL, dns.Type(ans.Type).String(), ans.Data)
			rr, err := dns.NewRR(record)
			if err != nil {
				log.Error().Str("module", "main").Str("record", record).Err(err).Send()
				panic(err)
			}
			m.Answer = append(m.Answer, rr)
		}
	}
}

///

func initConfig() *config.Config {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	port := flag.Int("port", 0, "DNS server port.")
	configFile := flag.String("conf", "", "Path to config file.")
	logLevel := flag.String("log-level", "", "Log level. debug, info, error")
	flag.Parse()

	cfg := new(config.Config)

	if len(*configFile) > 0 {
		cfg.Load(*configFile)
	}

	if *port != 0 {
		cfg.Port = *port
	}
	if cfg.Port == 0 {
		panic("'0' is not a valid port number")
	}

	if len(*logLevel) > 0 {
		zerolog.SetGlobalLevel(string2level(*logLevel))
	} else if len(cfg.LogLevel) > 0 {
		zerolog.SetGlobalLevel(string2level(cfg.LogLevel))
	}

	return cfg
}

func string2level(s string) zerolog.Level {
	switch s {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		panic("invalid log level: " + s)
	}
}
