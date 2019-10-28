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
	mc.GET("/", GetAllContainers)
	mc.POST("/", CreateContainer)
	mc.GET("/:id", GetContainer)
	mc.DELETE("/:id", DeleteContainer)
	mc.PUT("/:id/link-model", LinkModel)

	// Model routes
	mm := mg.Group("/models")
	mm.GET("/", GetAllModels)
	mm.POST("/", CreateModel)
	mm.GET("/:id", GetModel)
	mm.DELETE("/:id", EmptyModel)
	mm.GET("/preview", GetDataPreview)
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
