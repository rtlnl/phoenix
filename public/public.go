package public

import (
	"github.com/rtlnl/phoenix/middleware"

	"github.com/gin-gonic/gin"
)

// Public is the struct that will retain the server for ingesting the
// event from the trackers
type Public struct {
	App *gin.Engine
}

// NewPublicAPI creates a new object holding the Gin Server
func NewPublicAPI(dbHost, dbNamespace string, dbPort int) (*Public, error) {
	// Creates a router without any middleware by default
	r := gin.Default()

	r.RedirectTrailingSlash = true

	// middleware to inject Redis to all the routes for caching the client
	r.Use(middleware.Aerospike(dbHost, dbNamespace, dbPort))

	// Routes
	r.GET("/", LongVersion)
	r.GET("/recommend", Recommend)
	r.GET("/healthz", Healthz)

	// Docs
	r.Static("/docs", "docs/swagger-public")

	return &Public{
		App: r,
	}, nil
}

// ListenAndServe will start running the server
func (p *Public) ListenAndServe(addr string) error {
	return p.App.Run(addr)
}
