package server

import (
	"flag"
)

func (s *DnsServer) ParseArgs() {
	host := flag.String("host", "", "DNS server host. (default \"127.0.0.1\")")
	port := flag.Int("port", 0, "DNS server port.")
	configFile := flag.String("conf", "", "Path to config file.")
	logLevel := flag.String("log-level", "", "Log level. trace, debug, info, warn, error, fatal, panic. (default \"info\")")
	flag.Parse()

	if len(*configFile) > 0 {
		s.Config.LoadConfigFile(s.ctx, *configFile)
	}

	if len(*host) > 0 {
		s.Config.Host = *host
	}
	if s.Config.Host == "" {
		s.Config.Host = "127.0.0.1"
	}

	if *port != 0 {
		s.Config.Port = *port
	}

	if len(*logLevel) > 0 {
		s.Config.LogLevel = *logLevel
	}
	if s.Config.LogLevel == "" {
		s.Config.LogLevel = "info"
	}
}
