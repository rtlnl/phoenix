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
	// Creates a router without any middleware by default
	r := gin.New()

	// Global middleware
	r.Use(gin.Logger())

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())

	return &Internal{
		App: r,
	}, nil
}

// Run will initialize the server and will listen to the specified
// port from the config file
func (c *Internal) Run(host, dbHost, dbNamespace, redisAddr, redisPassword, s3Region, s3Endpoint string, s3DisableSSL bool, dbPort int) error {
	c.App.RedirectTrailingSlash = true

	// middlewares for injecting clients that are always in used. Caching is important when low latency is due
	c.App.Use(middleware.Aerospike(dbHost, dbNamespace, dbPort))
	c.App.Use(middleware.AWSSession(s3Region, s3Endpoint, s3DisableSSL))
	c.App.Use(middleware.Redis(redisAddr, redisPassword))

	// Routes
	c.App.GET("/", LongVersion)
	c.App.POST("/streaming", CreateStreaming)
	c.App.PUT("/streaming", UpdateStreaming)
	c.App.DELETE("/streaming", DeleteStreaming)
	c.App.POST("/batch", Batch)
	c.App.GET("/batch/status/:id", BatchStatus)

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
