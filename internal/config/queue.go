package config

type QueueConfig struct {
	Backend string `env:"QUEUE_BACKEND" envDefault:"postgres"`
}
