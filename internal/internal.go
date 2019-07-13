package internal

import (
	"github.com/gin-gonic/gin"
	"github.com/rtlnl/data-personalization-api/middleware"
)

// Internal is the struct that will retain the server for ingesting the
// event from the trackers
type Internal struct {
	App *gin.Engine
}

// NewInternalAPI creates a new Collector object
func NewInternalAPI() (*Internal, error) {
	return &Internal{
		App: gin.Default(),
	}, nil
}

// Run will initialize the server and will listen to the specified
// port from the config file
func (c *Internal) Run(host, dbHosts, dbPassword, s3Bucket, s3Region string) error {
	c.App.RedirectTrailingSlash = true

	// middlewares for injecting clients that are always in used. Caching is important when low latency is due
	c.App.Use(middleware.Redis(dbHosts, dbPassword))
	c.App.Use(middleware.S3(s3Bucket, s3Region))

	// Routes
	c.App.POST("/populate", Populate)

	return c.App.Run(host)
}
