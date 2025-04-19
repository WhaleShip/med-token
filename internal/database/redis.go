package database

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/whaleship/med-token/internal/config"
)

func GetInitRedis(cfg *config.Config) *redis.Client {
	rConn := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})

	if ping := rConn.Ping(context.Background()); ping.Err() != nil {
		log.Fatalln(ping)
	}
	return rConn
}
