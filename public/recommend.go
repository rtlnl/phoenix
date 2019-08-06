package public

import (
	"net/http"

	"github.com/rtlnl/data-personalization-api/models"

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
	ac := c.MustGet("AerospikeClient").(*db.AerospikeClient)

	var rr RecommendRequest
	if err := c.BindJSON(&rr); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	m, err := models.GetExistingModel(rr.PublicationPoint, rr.Campaign, ac)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	// compose key to retrieve recommended items
	// TODO: fix this
	key := m.ComposeSignalKey(rr.Signals[0])

	sn := m.ComposeSetName()
	r, err := ac.GetOne(sn, key)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	reccItems := r.Bins[key].([]string)

	utils.Response(c, http.StatusCreated, &RecommendResponse{
		Signals:         rr.Signals,
		Recommendations: reccItems,
	})
}
