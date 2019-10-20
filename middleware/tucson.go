package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rtlnl/phoenix/pkg/tucson"
)

// Tucson is a middleware to inject the Tucson client for accessing the gRPC service
func Tucson(address string) gin.HandlerFunc {
	client := tucson.NewClient(address)
	if err := client.Ping(); err != nil {
		panic(err)
	}
	return func(c *gin.Context) {
		c.Set("TucsonClient", client)
		c.Next()
	}
}
