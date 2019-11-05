package public

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rtlnl/phoenix/utils"
)

// Version is the current version of the APIs
const version = "v0.0.1"

// VersionResponse is the object that represents the payload of the root endpoint
type VersionResponse struct {
	Version string `json:"version"`
}

// LongVersion returns the current version of the APIs
func LongVersion(c *gin.Context) {
	v := fmt.Sprintf("Public Phoenix APIs %s", version)
	utils.Response(c, http.StatusOK, &VersionResponse{Version: v})
}
