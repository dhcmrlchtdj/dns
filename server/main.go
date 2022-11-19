package server

import (
	"context"
	"net"
	"os"
	"strconv"
	"sync"

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
	server.router.defaultRouter = new(routerNode)
	server.router.recordRouter = make(map[uint16]*routerNode)

	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	server.ctx = logger.WithContext(context.Background())

	return server
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
	s.cleanupExpiredCache()
	if err := s.dnsServer.ListenAndServe(); err != nil {
		zerolog.Ctx(s.ctx).
			Error().
			Str("module", "server.main").
			Err(err).
			Send()
		panic(err)
	}
}
