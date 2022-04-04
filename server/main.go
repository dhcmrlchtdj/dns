package server

import (
	"net"
	"strconv"
	"sync"

	"github.com/dhcmrlchtdj/godns/config"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type DnsServer struct {
	dnsServer *dns.Server
	config    config.Config
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
	s.router.addRules(s.config.Rule)
}

func (s *DnsServer) InitServer() {
	dnsMux := dns.NewServeMux()
	s.dnsServer = &dns.Server{
		Addr:    net.JoinHostPort(s.config.Host, strconv.Itoa(s.config.Port)),
		Net:     "udp",
		Handler: dnsMux,
		NotifyStartedFunc: func() {
			addr := s.dnsServer.PacketConn.LocalAddr()
			log.Info().
				Str("module", "server.main").
				Str("log_level", s.config.LogLevel).
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
