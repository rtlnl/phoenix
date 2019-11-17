package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rtlnl/phoenix/pkg/db"
)

// DB is a middleware to inject the Database client
func DB(d db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("DB", d)
		c.Next()
	}
}
