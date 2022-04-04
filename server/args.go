package server

import (
	"flag"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func (s *DnsServer) ParseArgs() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	host := flag.String("host", "", "DNS server host. (default \"127.0.0.1\")")
	port := flag.Int("port", 0, "DNS server port.")
	configFile := flag.String("conf", "", "Path to config file.")
	logLevel := flag.String("log-level", "", "Log level. trace, debug, info, error. (default \"info\")")
	flag.Parse()

	if len(*configFile) > 0 {
		s.config.LoadConfigFile(*configFile)
	}

	if len(*host) > 0 {
		s.config.Host = *host
	}
	if s.config.Host == "" {
		s.config.Host = "127.0.0.1"
	}

	if *port != 0 {
		s.config.Port = *port
	}
	if s.config.Port == 0 {
		panic("DNS server port is required")
	}

	if len(*logLevel) > 0 {
		s.config.LogLevel = *logLevel
	}
	if s.config.LogLevel == "" {
		s.config.LogLevel = "info"
	}
	zerolog.SetGlobalLevel(string2level(s.config.LogLevel))
}

func string2level(s string) zerolog.Level {
	switch s {
	case "trace":
		return zerolog.TraceLevel
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
