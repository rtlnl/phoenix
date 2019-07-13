package internal

import (
	"net/http"

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
func (c *Internal) Run(host, dbURL, dbPort, dbUsername, dbPassword, dbName string) error {
	c.App.RedirectTrailingSlash = true

	// middleware to inject Redis to all the routes for caching the client
	c.App.Use(middleware.Redis(dbURL, dbPort, dbUsername, dbPassword, dbName))

	// Routes
	c.App.GET("/populate", populate)

	return c.App.Run(host)
}

// populate will take care of populating the personalized content for all the users
func populate(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, `{"message":"populated"}`)
}
