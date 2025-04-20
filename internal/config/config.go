package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port          string
	RedisAddr     string
	RedisPassword string
	JWTSecret     []byte
	SMTPHost      string
	SMTPPort      int
	SMTPUsername  string
	SMTPPassword  string
}

func Load() *Config {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	return &Config{
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		JWTSecret:     []byte(os.Getenv("JWT_SECRET")),
		SMTPHost:      os.Getenv("SMTP_HOST"),
		SMTPPort:      port,
		SMTPUsername:  os.Getenv("SMTP_USER"),
		SMTPPassword:  os.Getenv("SMTP_PASS"),
	}
}
