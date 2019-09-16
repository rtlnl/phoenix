package public

import (
	"github.com/rtlnl/data-personalization-api/middleware"

	"github.com/gin-gonic/gin"
)

// Public is the struct that will retain the server for ingesting the
// event from the trackers
// gin is a framework, read about it
// Kind of a class: This struct allows us to declare new objects to compose your class (composition vs inheritance)
type Public struct {
	App *gin.Engine
}

// NewPublicAPI creates a new Collector object
// This is kind of a constructor where amongst others we initialize
// Always returns an error because
func NewPublicAPI() (*Public, error) {
	// Creates a router without any middleware by default
	r := gin.New()

	// Global middleware
	r.Use(gin.Logger())

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())

	return &Public{
		App: r,
	}, nil
}

// Run will initialize the server and will listen to the specified
// port from the config file
// We protect it by not allowing anything different than Public
// Instantiate in a lazy-loading way to optimize the usage of resources as well as to spin them up
func (c *Public) Run(host, dbHost, dbNamespace string, dbPort int) error {

	// Add if misses the last slash
	c.App.RedirectTrailingSlash = true

	// middleware to inject Aerospike to all the routes for caching the client
	// create globally to avoid multiple connections
	c.App.Use(middleware.Aerospike(dbHost, dbNamespace, dbPort))

	// Routes
	c.App.GET("/", LongVersion) // To prevent a 404 when calling an empty route
	c.App.POST("/recommend", Recommend)

	// Healthz
	// Kubernetes naming connection
	c.App.GET("/healthz", Healthz)

	// Docs
	c.App.Static("/docs", "docs/swagger-public")

	return c.App.Run(host)
}
