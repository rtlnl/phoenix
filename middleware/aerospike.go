package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rtlnl/data-personalization-api/pkg/db"
)

// Global middleware is loaded at the beginning (not lazy loaded)
// whilst route middleware are lazy loaded

// Aerospike is a middleware to inject the Aerospike client for accessing the database
// returns HandlerFunc is global so it's accesible everywhere
func Aerospike(dbHost, dbNamespace string, dbPort int) gin.HandlerFunc {
	client := db.NewAerospikeClient(dbHost, dbNamespace, dbPort)

	// check if aerospike is healthy
	if err := client.Health(); err != nil {
		panic(err) // system.exit(): crashes because w/o DB there's nothing else to do
	}

	return func(c *gin.Context) {
		c.Set("AerospikeClient", client)
		c.Next() // allows the request to be passed to the next stage
	}
}
