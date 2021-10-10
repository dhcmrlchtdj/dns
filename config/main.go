package config

import (
	"encoding/json"
	"os"

	"github.com/rs/zerolog/log"
)

///

type Config struct {
	Host     string   `json:"host,omitempty"`
	Port     int      `json:"port,omitempty"`
	LogLevel string   `json:"logLevel,omitempty"`
	Forward  []Server `json:"forward"`
}

type Server struct {
	DNS        string   `json:"dns"`
	HttpsProxy string   `json:"https_proxy,omitempty"`
	Domain     []string `json:"domain"`
}

///

func (c *Config) Load(file string) {
	log.Info().Str("module", "config").Str("path", file).Msg("load config")

	f, err := os.Open(file)
	if err != nil {
		log.Error().Str("module", "config").Str("path", file).Err(err).Send()
		panic(err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	dec.UseNumber()
	if err := dec.Decode(c); err != nil {
		log.Error().Str("module", "config").Str("path", file).Err(err).Send()
		panic(err)
	}
}
