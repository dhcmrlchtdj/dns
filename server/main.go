package server

import (
	"net"
	"strconv"
	"sync"

	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"

	"github.com/dhcmrlchtdj/godns/config"
)

type DnsServer struct {
	dnsServer *dns.Server
	Config    config.Config
	router    router
	cache     sync.Map
}

func NewDnsServer() *DnsServer {
	server := new(DnsServer)
	server.router.defaultRouter = new(routerNode)
	server.router.recordRouter = make(map[uint16]*routerNode)
	return server
}

func (s *DnsServer) InitRouter() {
	log.Debug().
		Str("module", "server.main").
		Msg("loading config")
	s.router.addRules(s.Config.Rule)
}

func (s *DnsServer) InitServer() {
	dnsMux := dns.NewServeMux()
	s.dnsServer = &dns.Server{
		Addr:    net.JoinHostPort(s.Config.Host, strconv.Itoa(s.Config.Port)),
		Net:     "udp",
		Handler: dnsMux,
		NotifyStartedFunc: func() {
			addr := s.dnsServer.PacketConn.LocalAddr()
			log.Info().
				Str("module", "server.main").
				Str("log_level", s.Config.LogLevel).
				Str("server_addr", addr.String()).
				Msg("DNS server is running")
		},
	}
	dnsMux.HandleFunc(".", s.handleRequest)
}

func (s *DnsServer) Start() {
	if err := s.dnsServer.ListenAndServe(); err != nil {
		log.Error().
			Str("module", "server.main").
			Err(err).
			Send()
		panic(err)
	}
}
