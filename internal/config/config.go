package config

import (
	"os"
)

type Config struct {
	RedisAddr     string
	RedisPassword string
}

func Load() *Config {
	return &Config{
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
	}
}
