package config

import (
	"encoding/json"
	"os"
)

///

type Config struct {
	Forward []Server `json:"forward"`
}

type Server struct {
	DNS    string   `json:"dns"`
	Domain []string `json:"domain,omitempty"`
}

///

func (c *Config) Load(file string) {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	dec.UseNumber()
	if err := dec.Decode(c); err != nil {
		panic(err)
	}
}
