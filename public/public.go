package public

import (
	"net/http"

	"github.com/rtlnl/data-personalization-api/middleware"

	"github.com/gin-gonic/gin"
)

// Public is the struct that will retain the server for ingesting the
// event from the trackers
type Public struct {
	App *gin.Engine
}

// NewPublicAPI creates a new Collector object
func NewPublicAPI() (*Public, error) {
	return &Public{
		App: gin.Default(),
	}, nil
}

// Run will initialize the server and will listen to the specified
// port from the config file
func (c *Public) Run(host, dbURL, dbPort, dbUsername, dbPassword, dbName string) error {
	c.App.RedirectTrailingSlash = true

	// middleware to inject Redis to all the routes for caching the client
	c.App.Use(middleware.Redis(dbURL, dbPort, dbUsername, dbPassword, dbName))

	// Routes
	c.App.GET("/recommend", recommend)

	return c.App.Run(host)
}

// recommend will take care of fetching the personalized content for a specific user
func recommend(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, `{"message":"recommended"}`)
}
