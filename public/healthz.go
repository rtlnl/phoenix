package public

import (
	"net/http"

	"github.com/rtlnl/data-personalization-api/utils"

	"github.com/gin-gonic/gin"
)

// HealthzResponse is the object that represents the response for the healthz endpoint
type HealthzResponse struct {
	Message string `json:"message"`
}

// Healthz will ping the api to make sure that it's alive
func Healthz(c *gin.Context) {
	utils.Response(c, http.StatusOK, &HealthzResponse{Message: "I'm healthy"})
}
