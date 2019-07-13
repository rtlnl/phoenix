package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rtlnl/data-personalization-api/pkg/db"
)

// Redis is a middleware to inject the Redis client for accessing the database
func Redis(dbHosts, dbPassword string) gin.HandlerFunc {
	addrs := strings.Split(dbHosts, ",")
	client := db.NewRedisClient(addrs, dbPassword)

	return func(c *gin.Context) {
		c.Set("RedisClient", client)
		c.Next()
	}
}
