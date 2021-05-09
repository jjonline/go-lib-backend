package client

import (
	"github.com/go-redis/redis/v7"
	"time"
)

// newRedis redis sample client
func NewRedis() *redis.Client {
	r := redis.NewClient(&redis.Options{
		Network:            "tcp",
		Addr:               "127.0.0.1:6379",
		Password:           "",
		DB:                 1,
		DialTimeout:        3 * time.Second,
		ReadTimeout:        3 * time.Second,
		WriteTimeout:       3 * time.Second,
		PoolSize:          3,
		MinIdleConns:       1,
		IdleTimeout:        60 * time.Second,
		IdleCheckFrequency: 10 * time.Minute,
	})

	return r
}
