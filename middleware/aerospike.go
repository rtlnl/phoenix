package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rtlnl/phoenix/pkg/db"
)

// Aerospike is a middleware to inject the Aerospike client for accessing the database
func Aerospike(dbHost, dbNamespace string, dbPort int) gin.HandlerFunc {
	client := db.NewAerospikeClient(dbHost, dbNamespace, dbPort)

	// check if aerospike is healthy
	if err := client.Health(); err != nil {
		panic(err)
	}

	return func(c *gin.Context) {
		c.Set("AerospikeClient", client)
		c.Next()
	}
}
