package config

import (
	"encoding/json"
	"os"

	"github.com/rs/zerolog/log"
)

type Config struct {
	Host     string  `json:"host,omitempty"`
	Port     int     `json:"port,omitempty"`
	LogLevel string  `json:"log_level,omitempty"`
	Rule     []*Rule `json:"rule,omitempty"`
}

type Rule struct {
	Pattern  Pattern  `json:"pattern"`
	Upstream Upstream `json:"upstream"`
}

type Pattern struct {
	Record string   `json:"record,omitempty"`
	Domain []string `json:"domain,omitempty"`
	Suffix []string `json:"suffix,omitempty"`
}

type Upstream struct {
	Block    string `json:"block,omitempty"`
	Ipv4     string `json:"ipv4,omitempty"`
	Ipv6     string `json:"ipv6,omitempty"`
	Udp      string `json:"udp,omitempty"`
	Doh      string `json:"doh,omitempty"`
	DohProxy string `json:"doh_proxy,omitempty"`
}

func (c *Config) LoadConfigFile(file string) {
	log.Info().Str("module", "config").Str("path", file).Msg("load config")

	f, err := os.Open(file)
	if err != nil {
		log.Error().Str("module", "server.config").Str("path", file).Err(err).Send()
		panic(err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	dec.UseNumber()
	if err := dec.Decode(c); err != nil {
		log.Error().Str("module", "config").Str("path", file).Err(err).Send()
		panic(err)
	}

	if len(c.Rule) > 0 {
		for _, rule := range c.Rule {
			if err := rule.IsValid(); err != nil {
				log.Error().Str("module", "config").Str("path", file).Err(err).Send()
				panic(err)
			}
		}
	}
}
