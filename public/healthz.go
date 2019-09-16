package public

import (
	"net/http"

	"github.com/rtlnl/data-personalization-api/utils"

	"github.com/gin-gonic/gin"
)

// HealthzResponse is the object that represents the response for the healthz endpoint
// marshall and unmarshall: 2 functions to create an object from string (and viceversa)
type HealthzResponse struct {
	Message string `json:"message"`
}

// Healthz will ping the api to make sure that it's alive
// You must set the gin interface for the GET call
func Healthz(c *gin.Context) {
	utils.Response(c, http.StatusOK, &HealthzResponse{Message: "I'm healthy"})
}
