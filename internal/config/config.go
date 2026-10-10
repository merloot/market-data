package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Database DatabaseConfig
	HTTP     HTTPConfig
	Redis    RedisConfig
	Queue    QueueConfig
	Realtime RealtimeConfig
	Provider ProviderConfig
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("Parse env: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if err := c.Database.Validate(); err != nil {
		return fmt.Errorf("Database: %w", err)
	}
	return nil
}
