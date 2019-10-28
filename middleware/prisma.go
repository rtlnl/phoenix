package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rtlnl/phoenix/pkg/prisma"
)

// PrismaClient injects the prisma client into the gin context
func PrismaClient(endpoint string) gin.HandlerFunc {
	client := prisma.New(&prisma.Options{Endpoint: endpoint})
	return func(c *gin.Context) {
		c.Set("PrismaClient", client)
		c.Next()
	}
}
