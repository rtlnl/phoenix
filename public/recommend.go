package public

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rtlnl/data-personalization-api/pkg/db"
	"github.com/rtlnl/data-personalization-api/utils"
)

// RecommendRequest is the object that represents the payload of the request for the recommend endpoint
type RecommendRequest struct {
	Signals          []Signal
	PublicationPoint string
	Campaign         string
}

// RecommendResponse is the object that represents the payload of the response for the recommend endpoint
type RecommendResponse struct {
	Signals         []Signal
	Recommendations []string
}

// Signal is an alias that represents the signal defintion
type Signal map[string]string

// Recommend will take care of fetching the personalized content for a specific user
func Recommend(c *gin.Context) {
	_ = c.MustGet("AerospikeClient").(*db.AerospikeClient)

	var rr RecommendRequest
	if err := c.BindJSON(&rr); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	// retrieve how the signal_key is composed
	// Example:
	//   - key: namespace#publicationPoint#campaign
	//	   value: articleId_userId

	// compose signal_key based on the value retrieved before

	// get recommended items

	utils.Response(c, http.StatusCreated, "recommended items")
}
