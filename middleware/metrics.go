package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rtlnl/phoenix/pkg/metrics"
)

// Metrics is the middleware that will inject the metrics client
func Metrics(mc metrics.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("MetricsClient", mc)
		c.Next()
	}
}
