package config

import (
	"fmt"
	"time"
)

type DatabaseConfig struct {
	URL               string        `env:"DB_URL,required"`
	MaxConns          int32         `env:"DB_MAX_CONNS" envDefault:"10"`
	MinConns          int32         `env:"DB_MIN_CONS" envDefault:"2"`
	MaxConnLifetime   time.Duration `env:"MAX_CONN_LIFETIME" envDefault:"1h"`
	MaxConnIdleTime   time.Duration `env:"MAX_CONN_IDLE_TIME" envDefault:"30m"`
	ConnectionTimeout time.Duration `env:"DB_CONNECTION_TIMEOUT" envDefault:"5s"`
}

func (c *DatabaseConfig) Validate() error {
	if c.MaxConns < c.MinConns {
		return fmt.Errorf("MaxConns(%d) must be >= MinConns(%d)", c.MaxConns, c.MinConns)
	}
	return nil
}