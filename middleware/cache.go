package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/rtlnl/phoenix/pkg/cache"
)

// Cache is a middleware instantiating the caching layer
func Cache(ch cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("CacheClient", ch)
		c.Next()
	}
}
