// Package config provides application configuration loaded from environment variables.
package config

import "github.com/kelseyhightower/envconfig"

// ServiceConfig holds HTTP server configuration
type ServiceConfig struct {
	BindAddress string `envconfig:"BIND_ADDRESS" default:"0.0.0.0:8080"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
}

// DBConfig holds database configuration
type DBConfig struct {
	Type     string `envconfig:"DB_TYPE" default:"pgsql"`
	Hostname string `envconfig:"DB_HOST" default:"localhost"`
	Port     string `envconfig:"DB_PORT" default:"5432"`
	Name     string `envconfig:"DB_NAME" default:"catalog-manager"`
	User     string `envconfig:"DB_USER" default:"admin"`
	Password string `envconfig:"DB_PASSWORD" default:"adminpass"`
}

// PlacementConfig holds Placement Manager configuration
type PlacementConfig struct {
	URL string `envconfig:"PLACEMENT_MANAGER_URL" default:"http://localhost:8081"`
}

// PetClinicConfig holds Pet Clinic catalog item configuration
type PetClinicConfig struct {
	RegionDefault string   `envconfig:"PETCLINIC_REGION_DEFAULT" default:""`
	RegionEnum    []string `envconfig:"PETCLINIC_REGION_ENUM" default:"region-a,region-b"`
}

func DefaultPetClinicConfig() PetClinicConfig {
	return PetClinicConfig{
		RegionEnum: []string{"region-a", "region-b"},
	}
}

// Config holds all configuration for the application
type Config struct {
	Service   ServiceConfig
	Database  DBConfig
	Placement PlacementConfig
	PetClinic PetClinicConfig
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg.Service); err != nil {
		return nil, err
	}
	if err := envconfig.Process("", &cfg.Database); err != nil {
		return nil, err
	}
	if err := envconfig.Process("", &cfg.Placement); err != nil {
		return nil, err
	}
	if err := envconfig.Process("", &cfg.PetClinic); err != nil {
		return nil, err
	}
	return &cfg, nil
}
