package db

import (
	"github.com/go-redis/redis"
)

// RedisClient is a wrapper around the official package
type RedisClient struct {
	Client *redis.Client
}

// NewRedisClient connects and return a redis client instance where to store/read information
func NewRedisClient(addr, username, password string, db int) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, // no password set
		DB:       db,       // use default DB
	})
	return &RedisClient{Client: client}
}
