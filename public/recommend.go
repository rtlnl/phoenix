package public

import (
	"errors"
	"net/http"
	"sync"

	"github.com/rtlnl/data-personalization-api/models"

	"github.com/gin-gonic/gin"
	"github.com/rtlnl/data-personalization-api/pkg/db"
	"github.com/rtlnl/data-personalization-api/utils"
)

const (
	signalSeparator = "_"
	binKey          = "items"
)

// RecommendRequest is the object that represents the payload of the request for the recommend endpoint
type RecommendRequest struct {
	Signals          []Signal `json:"signals" binding:"required"`
	PublicationPoint string   `json:"publicationPoint" binding:"required"`
	Campaign         string   `json:"campaign" binding:"required"`
}

// RecommendResponse is the object that represents the payload of the response for the recommend endpoint
type RecommendResponse struct {
	Signals         []Signal    `json:"signals"`
	Recommendations interface{} `json:"recommendations"`
}

// Signal is an alias that represents the signal defintion
type Signal map[string]string

// rrPool is in charged of Pooling eventual requests in coming. This will help to reduce the alloc/s
// and effeciently improve the garbage collection operations. rr is short for recommend-request
var rrPool = sync.Pool{
	New: func() interface{} { return new(RecommendRequest) },
}

// Recommend will take care of fetching the personalized content for a specific user
func Recommend(c *gin.Context) {
	ac := c.MustGet("AerospikeClient").(*db.AerospikeClient)

	// get a new object from the pool and then dispose it
	rr := rrPool.Get().(*RecommendRequest)
	defer rrPool.Put(rr)

	if err := c.BindJSON(rr); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	m, err := models.GetExistingModel(rr.PublicationPoint, rr.Campaign, ac)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}

	// If the model is staged, the clients cannot access it
	if m.IsStaged() {
		utils.ResponseError(c, http.StatusBadRequest, errors.New("model is staged. Clients cannot access staged models"))
		return
	}

	// compose key to retrieve recommended items
	ss := make(map[string]string, len(rr.Signals))
	for _, s := range rr.Signals {
		for k, v := range s {
			ss[k] = v
		}
	}

	key := m.ComposeSignalKey(ss, signalSeparator)
	if key == "" {
		utils.ResponseError(c, http.StatusBadRequest, errors.New("signal is not formatted correctly"))
		return
	}

	sn := m.ComposeSetName()
	r, err := ac.GetOne(sn, key)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}

	utils.Response(c, http.StatusOK, &RecommendResponse{
		Signals:         rr.Signals,
		Recommendations: r.Bins[binKey],
	})
}
