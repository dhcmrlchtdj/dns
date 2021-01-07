package config

///

func Read() {
}


///

type Config struct {
	Server []ServerConfig `json:"server"`
}

type ServerConfig struct {
	DNSType   string   `json:"dns_type"`
	DNSServer string   `json:"dns_server"`
	Default   bool     `json:"default,omitempty"`
	Domain    []string `json:"domain,omitempty"`
}
