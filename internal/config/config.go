package config

import "github.com/kelseyhightower/envconfig"

// ServiceConfig holds HTTP server configuration
type ServiceConfig struct {
	BindAddress string `envconfig:"BIND_ADDRESS" default:"0.0.0.0:8080"`
}

// DBConfig holds database configuration
type DBConfig struct {
	Type     string `envconfig:"DB_TYPE" default:"sqlite"`
	Hostname string `envconfig:"DB_HOST" default:"localhost"`
	Port     string `envconfig:"DB_PORT" default:"5432"`
	Name     string `envconfig:"DB_NAME" default:"catalog-manager.db"`
	User     string `envconfig:"DB_USER" default:""`
	Password string `envconfig:"DB_PASSWORD" default:""`
}

// Config holds all configuration for the application
type Config struct {
	Service  ServiceConfig
	Database DBConfig
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg.Service); err != nil {
		return nil, err
	}
	if err := envconfig.Process("", &cfg.Database); err != nil {
		return nil, err
	}
	return &cfg, nil
}
