package internal

import (
	"github.com/gin-gonic/gin"
)

// Internal is the struct that will retain the server for ingesting the
// event from the trackers
type Internal struct {
	App *gin.Engine
}

// NewInternalAPI creates the o object
func NewInternalAPI(middlewares ...gin.HandlerFunc) (*Internal, error) {
	// Creates a router without any middleware by default
	r := gin.Default()

	r.RedirectTrailingSlash = true

	// add all middleware
	for _, m := range middlewares {
		r.Use(m)
	}

	// Internal API v1
	v1 := r.Group("v1")

	// Routes
	v1.GET("/", LongVersion)

	v1.POST("/streaming", CreateStreaming)
	v1.PUT("/streaming", UpdateStreaming)
	v1.DELETE("/streaming", DeleteStreaming)
	v1.POST("/batch", Batch)
	v1.GET("/batch/status/:id", BatchStatus)

	// Management Routes
	mg := v1.Group("/management")

	// Container routes
	mc := mg.Group("/containers")
	mc.GET("/", GetContainer)
	mc.POST("/", CreateContainer)
	mc.DELETE("/", EmptyContainer)
	mc.GET("/all", GetAllContainers)
	mc.PUT("/link-model", LinkModel)

	// Model routes
	mm := mg.Group("/models")
	mm.GET("/", GetModel)
	mm.POST("/", CreateModel)
	mm.DELETE("/", EmptyModel)
	mm.GET("/preview", GetDataPreview)
	mm.GET("/all", GetAllModels)
	mm.POST("/publish", PublishModel)
	mm.POST("/stage", StageModel)

	// Healthz
	v1.GET("/healthz", Healthz)

	// Docs
	v1.Static("/docs", "docs/swagger-internal")

	return &Internal{
		App: r,
	}, nil
}

// ListenAndServe will initialize the server and will listen to the specified
// port from the config file
func (i *Internal) ListenAndServe(host string) error {
	return i.App.Run(host)
}
