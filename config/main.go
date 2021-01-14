package config

///

func Read() {
}

///

type Config struct {
	Server []ServerConfig `json:"server"`
}

type ServerConfig struct {
	DNS    string   `json:"dns"`
	Domain []string `json:"domain,omitempty"`
}
