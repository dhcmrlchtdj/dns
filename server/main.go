package server

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	_ "net/http/pprof"

	"github.com/miekg/dns"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/dhcmrlchtdj/godns/config"
)

type DnsServer struct {
	dnsServer     *dns.Server
	pprofServer   *http.Server
	pprofListener net.Listener
	ctx           context.Context
	router        router
	cache         sync.Map
	Config        config.Config
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

		s.shutdownDNS()
		s.shutdownPprof()
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

func (s *DnsServer) SetupPprof() {
	pprofMux := http.DefaultServeMux
	http.DefaultServeMux = http.NewServeMux()
	s.pprofServer = &http.Server{
		Handler: pprofMux,
		BaseContext: func(_ net.Listener) context.Context {
			return s.ctx
		},
	}
	netListener, err := net.Listen("tcp", s.Config.Host+":0")
	if err != nil {
		zerolog.Ctx(s.ctx).
			Error().
			Str("module", "server.main").
			Stack().
			Err(err).
			Send()
		panic(err)
	}
	s.pprofListener = netListener
}

func (s *DnsServer) Start() {
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		s.cleanupExpiredCache()
		wg.Done()
	}()

	go func() {
		s.startDNS()
		wg.Done()
	}()

	go func() {
		s.startPprof()
		wg.Done()
	}()

	wg.Wait()
}

func (s *DnsServer) startDNS() {
	err := s.dnsServer.ListenAndServe()
	if err != nil {
		zerolog.Ctx(s.ctx).
			Error().
			Str("module", "server.main").
			Stack().
			Err(err).
			Send()
		panic(err)
	}
}

func (s *DnsServer) shutdownDNS() {
	err := s.dnsServer.Shutdown()
	if err != nil {
		zerolog.Ctx(s.ctx).
			Error().
			Str("module", "server.main").
			Stack().
			Err(err).
			Send()
		panic(err)
	}
}

func (s *DnsServer) startPprof() {
	zerolog.Ctx(s.ctx).
		Info().
		Str("module", "server.main").
		Str("pprof_addr", "http://"+s.pprofListener.Addr().String()+"/debug/pprof/").
		Msg("pprof is running")

	err := s.pprofServer.Serve(s.pprofListener)
	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			zerolog.Ctx(s.ctx).
				Error().
				Str("module", "server.main").
				Stack().
				Err(err).
				Send()
			panic(err)
		}
	}
}

func (s *DnsServer) shutdownPprof() {
	err := s.pprofServer.Shutdown(s.ctx)
	if err != nil {
		zerolog.Ctx(s.ctx).
			Error().
			Str("module", "server.main").
			Stack().
			Err(err).
			Send()
		panic(err)
	}
}
