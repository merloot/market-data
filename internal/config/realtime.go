package config

type RealtimeConfig struct {
	Backend string `env:"REALTIME_BACKEND" envDefault:"memory"`
}
