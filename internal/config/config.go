package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	BindAddress string `envconfig:"BIND_ADDRESS" default:"0.0.0.0:8080"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
