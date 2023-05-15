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
	Clients     []string `yaml:"clients"`
	MaxFileSize int64    `yaml:"maxFileSize"`
	LocalPath   string   `yaml:"backupPathLocal"`
	RemotePath  string   `yaml:"backupPathIntermediateService"`
}

func LoadConfig(file string) (Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return Config{}, fmt.Errorf("error loading config file: %w", err)
	}

	var cfg Config
	err = yaml.NewDecoder(f).Decode(&cfg)
	if cfg.MaxFileSize == 0 {
		cfg.MaxFileSize = 100 * 1024 * 1024
	} else {
		cfg.MaxFileSize = cfg.MaxFileSize * 1024 * 1024
	}

	if err != nil {
		err = fmt.Errorf("error decoding config file: %w", err)
	}
	return cfg, err
}
