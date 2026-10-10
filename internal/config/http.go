package config

import "time"

type HTTPConfig struct {
	Addr            string        `env:"HTTP_ADDR" envDefault:"8080"`
	ReadTimeout     time.Duration `env:"HTTP_TIME_DURATION" envDefault:"10s"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"30s"`
	ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" envDefault:"15s"`
}
