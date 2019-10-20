package public

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/rtlnl/phoenix/models"

	"github.com/gin-gonic/gin"
	"github.com/rtlnl/phoenix/pkg/db"
	"github.com/rtlnl/phoenix/pkg/tucson"
	"github.com/rtlnl/phoenix/utils"
)

const (
	binKey = "items"
)

// RecommendRequest is the object that represents the payload of the request for the recommend endpoint
type RecommendRequest struct {
	SignalID         string
	PublicationPoint string
	Campaign         string
	ModelName        string
}

// RecommendResponse is the object that represents the payload of the response for the recommend endpoint
type RecommendResponse struct {
	Recommendations interface{} `json:"recommendations"`
}

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

	// get query parameters from URL
	pp := c.DefaultQuery("publicationPoint", "")
	cp := c.DefaultQuery("campaign", "")
	mn := c.DefaultQuery("model", "")
	sID := c.DefaultQuery("signalId", "")

	if err := validateRecommendQueryParameters(rr, pp, cp, mn, sID); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	m, err := models.GetExistingModel(rr.PublicationPoint, rr.Campaign, rr.ModelName, ac)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}

	// If the model is staged, the clients cannot access it
	if m.IsStaged() {
		utils.ResponseError(c, http.StatusBadRequest, errors.New("model is staged. Clients cannot access staged models"))
		return
	}

	// call Tucson here
	tc, exists := c.Get("TucsonClient")
	if exists {
		modelName, _ := tc.(*tucson.Client).GetModel(pp, cp)
		log.Info().Msgf("model name %s", modelName)
	}

	sn := m.ComposeSetName()
	r, err := ac.GetOne(sn, rr.SignalID)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}

	// convert single entry from interface{} to []models.ItemScore
	itemsScore := convertSingleEntry(r.Bins[binKey])

	utils.Response(c, http.StatusOK, &RecommendResponse{
		Recommendations: itemsScore,
	})
}

func validateRecommendQueryParameters(rr *RecommendRequest, publicationPoint, campaign, modelName, signalID string) error {
	var mp []string

	// TODO: improve this in somehow
	if publicationPoint == "" {
		mp = append(mp, "publicationPoint")
	}

	if campaign == "" {
		mp = append(mp, "campaign")
	}

	if modelName == "" {
		mp = append(mp, "model")
	}

	if signalID == "" {
		mp = append(mp, "signalId")
	}

	if len(mp) > 0 {
		return fmt.Errorf("missing %s in the URL query", strings.Join(mp[:], ","))
	}

	// update values
	rr.PublicationPoint = publicationPoint
	rr.Campaign = campaign
	rr.ModelName = modelName
	rr.SignalID = signalID

	return nil
}

// The objects coming from Aerospike have type []interface{}. This function converts
// the Bins in the appropriate type for consistency
func convertSingleEntry(bins interface{}) []models.ItemScore {
	var itemsScore []models.ItemScore
	newBins := bins.([]interface{})
	for _, bin := range newBins {
		b := bin.(map[interface{}]interface{})
		item := make(models.ItemScore)
		for k, v := range b {
			it := fmt.Sprintf("%v", k)
			score := fmt.Sprintf("%v", v)
			item[it] = score
		}
		itemsScore = append(itemsScore, item)
	}
	return itemsScore
}
