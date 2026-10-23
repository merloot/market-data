package config

type RedisConfig struct {
	Addr string `env:"REDIS_ADDRESS"`
}
