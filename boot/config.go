package boot

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Addr string `yaml:"addr"`
	} `yaml:"server"`
	Clients []string `yaml:"clients"`
}

func LoadConfig(file string) (Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return Config{}, fmt.Errorf("error loading config file: %w", err)
	}

	var cfg Config
	err = yaml.NewDecoder(f).Decode(&cfg)
	if err != nil {
		err = fmt.Errorf("error decoding config file: %w", err)
	}
	return cfg, err
}
