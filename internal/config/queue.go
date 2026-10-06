package config

type QueueConfig struct {
	Backend string `env:"BACKEND_QUEUE" envDefault:"postgres"`
}
