package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rtlnl/phoenix/pkg/logs"
)

// RecommendationLogs is a middleware to inject the recommendation logger
func RecommendationLogs(r logs.RecommendationLog) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("RecommendationLog", r)
		c.Next()
	}
}
