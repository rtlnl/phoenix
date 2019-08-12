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
func (c *Internal) Run(host, dbHost, dbNamespace, s3Bucket, s3Region, s3Endpoint string, s3DisableSSL bool, dbPort int) error {
	c.App.RedirectTrailingSlash = true

	// middlewares for injecting clients that are always in used. Caching is important when low latency is due
	c.App.Use(middleware.Aerospike(dbHost, dbNamespace, dbPort))
	c.App.Use(middleware.S3(s3Bucket, s3Region, s3Endpoint, s3DisableSSL))

	// Routes
	c.App.GET("/", LongVersion)
	c.App.POST("/streaming", CreateStreaming)
	c.App.PUT("/streaming", UpdateStreaming)
	c.App.DELETE("/streaming", DeleteStreaming)
	c.App.POST("/batch", Batch)

	// Management Routes
	mm := c.App.Group("/management/model")
	mm.GET("/", GetModel)
	mm.POST("/", CreateModel)
	mm.DELETE("/", EmptyModel)
	mm.POST("/publish", PublishModel)
	mm.POST("/stage", StageModel)

	// Healthz
	c.App.GET("/healthz", Healthz)

	// Docs
	c.App.Static("/docs", "docs/swagger-internal")

	return c.App.Run(host)
}
