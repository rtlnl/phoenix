package middleware

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rtlnl/data-personalization-api/pkg/db"
)

// Redis is a middleware to inject the Redis client for accessing the database
func Redis(dbURL, dbPort, dbUsername, dbPassword, dbName string) gin.HandlerFunc {

	dbNameInt, err := strconv.Atoi(dbName)
	if err != nil {
		panic(err)
	}

	addr := fmt.Sprintf("%s:%s", dbURL, dbPort)
	client := db.NewRedisClient(addr, dbUsername, dbPassword, dbNameInt)

	return func(c *gin.Context) {
		c.Set("RedisClient", client)
		c.Next()
	}
}
