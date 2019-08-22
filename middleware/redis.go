package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

// Redis is a middleware to inject the Redis client for accessing the database
func Redis(addr, password string) gin.HandlerFunc {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, // no password set
		DB:       0,        // use default DB
	})

	// check if Redis is healthy
	if _, err := client.Ping().Result(); err != nil {
		panic(err)
	}

	return func(c *gin.Context) {
		c.Set("RedisClient", client)
		c.Next()
	}
}
