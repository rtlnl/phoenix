package internal

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rtlnl/data-personalization-api/utils"
)

// Version is the current version of the APIs
const version = "v0.0.1"

// LongVersion returns the current version of the APIs
func LongVersion(c *gin.Context) {
	v := fmt.Sprintf("Internal Personalization APIs %s", version)
	utils.Response(c, http.StatusOK, v)
}
