package server

import (
	"context"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/miekg/dns"
	"github.com/rs/zerolog"

	"github.com/dhcmrlchtdj/godns/config"
)

type DnsServer struct {
	dnsServer *dns.Server
	Config    config.Config
	router    router
	cache     sync.Map
	ctx       context.Context
}

func NewDnsServer() *DnsServer {
	server := new(DnsServer)
	server.router.setup()
	return server
}

func (s *DnsServer) SetupContext() {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	ctx, cancel := context.WithCancel(context.Background())
	s.ctx = logger.WithContext(ctx)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		sig := <-c
		logger.Info().
			Str("module", "server.main").
			Str("signal", sig.String()).
			Msg("DNS server is stopping")
		cancel()
		err := s.dnsServer.Shutdown()
		if err != nil {
			logger.Error().
				Str("module", "server.main").
				Stack().
				Err(err).
				Send()
			panic(err)
		}
	}()
}

func (s *DnsServer) SetupRouter() {
	zerolog.Ctx(s.ctx).
		Debug().
		Str("module", "server.main").
		Msg("loading config")
	s.router.addRules(s.ctx, s.Config.Rule)
}

func (s *DnsServer) SetupServer() {
	dnsMux := dns.NewServeMux()
	s.dnsServer = &dns.Server{
		Addr:    net.JoinHostPort(s.Config.Host, strconv.Itoa(s.Config.Port)),
		Net:     "udp",
		Handler: dnsMux,
		NotifyStartedFunc: func() {
			addr := s.dnsServer.PacketConn.LocalAddr()
			zerolog.Ctx(s.ctx).
				Info().
				Str("module", "server.main").
				Str("log_level", s.Config.LogLevel).
				Str("server_addr", addr.String()).
				Msg("DNS server is running")
		},
	}
	dnsMux.HandleFunc(".", s.handleRequest)
}

func (s *DnsServer) Start() {
	go s.cleanupExpiredCache()
	if err := s.dnsServer.ListenAndServe(); err != nil {
		zerolog.Ctx(s.ctx).
			Error().
			Str("module", "server.main").
			Stack().
			Err(err).
			Send()
		panic(err)
	}
}
