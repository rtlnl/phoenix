package public

import (
	"net/http"

	"github.com/rtlnl/data-personalization-api/utils"

	"github.com/gin-gonic/gin"
)

// Healthz will ping the api to make sure that it's alive
func Healthz(c *gin.Context) {
	utils.Response(c, http.StatusOK, "I'm healthy")
}
