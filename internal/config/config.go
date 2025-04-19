package config

import (
	"os"
)

type Config struct {
	RedisAddr     string
	RedisPassword string
	JWTSecret     []byte
}

func Load() *Config {
	return &Config{
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		JWTSecret:     []byte(os.Getenv("JWT_SECRET")),
	}
}
