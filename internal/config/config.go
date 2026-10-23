package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
	HTTP     HTTPConfig
	Provider ProviderConfig
	Realtime RealtimeConfig
	Queue    QueueConfig
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse env: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	return nil
}
