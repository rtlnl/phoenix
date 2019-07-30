package public

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rtlnl/data-personalization-api/utils"
)

// Recommend will take care of fetching the personalized content for a specific user
func Recommend(c *gin.Context) {
	utils.Response(c, http.StatusOK, "recommended")
}
