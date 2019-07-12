package redis

import (
	"fmt"

	"github.com/go-redis/redis"
)

// NewRedisClient connects and return a redis client instance where to store/read information
func NewRedisClient(addr, password string, db int) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, // no password set
		DB:       db,       // use default DB
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
	// Output: PONG <nil>
}
