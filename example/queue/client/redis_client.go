package client

import (
	"github.com/redis/go-redis/v9"
	"time"
)

// newRedis redis sample client
func NewRedis() *redis.Client {
	r := redis.NewClient(&redis.Options{
		Network:      "tcp",
		Addr:         "127.0.0.1:6379",
		Password:     "",
		DB:           1,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     3,
		MinIdleConns: 1,
	})

	return r
}
