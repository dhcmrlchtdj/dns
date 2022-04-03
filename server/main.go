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

func NewDnsServer() DnsServer {
	server := DnsServer{}
	server.router.defaultRouter = new(routerNode)
	server.router.recordRouter = make(map[uint16]*routerNode)
	return server
}

func (s *DnsServer) InitDnsRouter() {
	s.router.addRules(s.config.Rule)
}

func (s *DnsServer) InitDnsServer() {
	dnsMux := dns.NewServeMux()
	s.dnsServer = &dns.Server{
		Addr:    net.JoinHostPort(s.config.Host, strconv.Itoa(s.config.Port)),
		Net:     "udp",
		Handler: dnsMux,
	}
	dnsMux.HandleFunc(".", s.handleRequest)
}

func (s *DnsServer) ListenAndServe() error {
	log.Info().
		Str("module", "server.main").
		Str("host", s.config.Host).
		Int("port", s.config.Port).
		Str("log_level", s.config.LogLevel).
		Msg("DNS server is running")
	return s.dnsServer.ListenAndServe()
}

func (s *DnsServer) Shutdown() error {
	return s.dnsServer.Shutdown()
}
