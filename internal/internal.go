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

// NewInternalAPI creates the o object
func NewInternalAPI(dbHost, dbNamespace, s3Region, s3Endpoint string, s3DisableSSL bool, dbPort int) (*Internal, error) {
	// Creates a router without any middleware by default
	r := gin.Default()

	r.RedirectTrailingSlash = true

	// middlewares for injecting clients that are always in used. Caching is important when low latency is due
	r.Use(middleware.Aerospike(dbHost, dbNamespace, dbPort))
	r.Use(middleware.AWSSession(s3Region, s3Endpoint, s3DisableSSL))

	// Routes
	r.GET("/", LongVersion)
	r.POST("/streaming", CreateStreaming)
	r.PUT("/streaming", UpdateStreaming)
	r.DELETE("/streaming", DeleteStreaming)
	r.POST("/batch", Batch)
	r.GET("/batch/status/:id", BatchStatus)

	// Management Routes
	mg := r.Group("/management")

	// Container routes
	mc := mg.Group("/containers")
	mc.GET("/", GetContainer)
	mc.POST("/", CreateContainer)
	mc.DELETE("/", EmptyContainer)
	mc.PUT("/link-model", LinkModel)

	// Model routes
	mm := mg.Group("/models")
	mm.GET("/", GetModel)
	mm.POST("/", CreateModel)
	mm.DELETE("/", EmptyModel)
	mm.POST("/publish", PublishModel)
	mm.POST("/stage", StageModel)

	// Healthz
	r.GET("/healthz", Healthz)

	// Docs
	r.Static("/docs", "docs/swagger-internal")

	return &Internal{
		App: r,
	}, nil
}

// ListenAndServe will initialize the server and will listen to the specified
// port from the config file
func (i *Internal) ListenAndServe(host string) error {
	return i.App.Run(host)
}
