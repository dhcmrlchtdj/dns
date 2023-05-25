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

	if *configFile != "" {
		s.Config.LoadConfigFile(s.ctx, *configFile)
	}

	if *host != "" {
		s.Config.Host = *host
	}
	if s.Config.Host == "" {
		s.Config.Host = "127.0.0.1"
	}

	if *port != 0 {
		s.Config.Port = *port
	}

	if *logLevel != "" {
		s.Config.LogLevel = *logLevel
	}
	if s.Config.LogLevel == "" {
		s.Config.LogLevel = "info"
	}
}
