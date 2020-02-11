package internal

import (
	ginprometheus "github.com/banzaicloud/go-gin-prometheus"
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

	// metrics with prometheus
	p := ginprometheus.NewPrometheus("phoenix_internal", []string{})
	p.SetListenAddress(":9900")
	p.Use(r, "/metrics")

	// Base path
	r.GET("/", LongVersion)
	r.GET("/healthz", Healthz)

	// Internal API v1
	v1 := r.Group("v1")
	v1.POST("/batch", Batch)
	v1.GET("/batch/status/:id", BatchStatus)

	sc := v1.Group("/streaming")
	sc.POST("/", CreateStreaming)
	sc.PUT("/", UpdateStreaming)
	sc.DELETE("/", DeleteStreaming)
	sc.POST("/like", HandleLike)

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

	return &Internal{
		App: r,
	}, nil
}

// ListenAndServe will initialize the server and will listen to the specified
// port from the config file
func (i *Internal) ListenAndServe(host string) error {
	return i.App.Run(host)
}
