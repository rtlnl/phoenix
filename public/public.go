package public

import (
	"github.com/rtlnl/data-personalization-api/middleware"

	"github.com/gin-gonic/gin"
)

// Public is the struct that will retain the server for ingesting the
// event from the trackers
type Public struct {
	App *gin.Engine
}

// NewPublicAPI creates a new Collector object
func NewPublicAPI() (*Public, error) {
	return &Public{
		App: gin.Default(),
	}, nil
}

// Run will initialize the server and will listen to the specified
// port from the config file
func (c *Public) Run(host, dbHost, dbNamespace string, dbPort int) error {
	c.App.RedirectTrailingSlash = true

	// middleware to inject Redis to all the routes for caching the client
	c.App.Use(middleware.Aerospike(dbHost, dbNamespace, dbPort))

	// Routes
	c.App.POST("/recommend", Recommend)

	// Healthz
	c.App.GET("/healthz", Healthz)

	return c.App.Run(host)
}
