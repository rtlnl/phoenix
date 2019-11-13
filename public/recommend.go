package public

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	zerolog "github.com/rs/zerolog/log"

	"github.com/rtlnl/phoenix/models"
	"github.com/rtlnl/phoenix/pkg/cache"
	"github.com/rtlnl/phoenix/pkg/db"
	"github.com/rtlnl/phoenix/pkg/logs"
	"github.com/rtlnl/phoenix/pkg/tucson"
	"github.com/rtlnl/phoenix/utils"
)

const (
	binKey = "items"
)

// RecommendRequest is the object that represents the payload of the request for the recommend endpoint
type RecommendRequest struct {
	SignalID         string `json:"signalId"`
	PublicationPoint string `json:"publicationPoint"`
	Campaign         string `json:"campaign"`
}

// RecommendResponse is the object that represents the payload of the response for the recommend endpoint
type RecommendResponse struct {
	ModelName       string      `json:"modelName"`
	Recommendations interface{} `json:"recommendations" description:""`
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
	sID := c.DefaultQuery("signalId", "")

	// validate recommendation parameters
	if err := validateRecommendQueryParameters(rr, pp, cp, sID); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	// get container from Aerospike
	container, err := models.GetExistingContainer(rr.PublicationPoint, rr.Campaign, ac)
	if container == nil || err != nil {
		utils.ResponseError(c, http.StatusNotFound, fmt.Errorf("container with publication point %s and campaign %s is not found", pp, cp))
		return
	}

	// get model name either from Tucson or URL
	modelName, err := getModelName(c, container)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}

	// get model from aerospike
	m, err := models.GetExistingModel(modelName, ac)
	if m == nil || err != nil {
		utils.ResponseError(c, http.StatusNotFound, fmt.Errorf("model %s not found", modelName))
		return
	}

	// if the model is staged, the clients cannot access it
	if m.IsStaged() {
		utils.ResponseError(c, http.StatusBadRequest, errors.New("model is staged. Clients cannot access staged models"))
		return
	}

	// get caching layer client
	cc := c.MustGet("CacheClient").(cache.Cache)
	// get logging client
	lt := c.MustGet("RecommendationLog").(logs.RecommendationLog)

	// compose key for the cache
	key := fmt.Sprintf("%s#%s", modelName, rr.SignalID)
	// check if value is in cache
	if is, ok := cc.Get(key); ok {
		// write logs
		lt.Write(logs.RowLog{
			PublicationPoint: rr.PublicationPoint,
			Campaign:         rr.Campaign,
			SignalID:         rr.SignalID,
			ItemScores:       is,
		})
		// return response
		utils.Response(c, http.StatusOK, &RecommendResponse{
			ModelName:       modelName,
			Recommendations: is,
		})
		return
	}

	// get the recommended values
	r, err := ac.GetOne(modelName, rr.SignalID)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}

	// convert single entry from interface{} to []models.ItemScore
	itemsScore := convertSingleEntry(r.Bins[binKey])

	// store in cache
	if ok := cc.Set(key, itemsScore); !ok {
		// if an error occur we simply log it and continue
		log.Error().Msgf("failed to store key %s in cache", key)
	}

	// write logs to the logging system
	lt.Write(logs.RowLog{
		PublicationPoint: rr.PublicationPoint,
		Campaign:         rr.Campaign,
		SignalID:         rr.SignalID,
		ItemScores:       itemsScore,
	})

	utils.Response(c, http.StatusOK, &RecommendResponse{
		ModelName:       modelName,
		Recommendations: itemsScore,
	})
}

func getModelName(c *gin.Context, container *models.Container) (string, error) {
	// check tucson
	modelName := getModelFromTucson(c, container.PublicationPoint, container.Campaign)
	if !utils.IsStringEmpty(modelName) {
		return modelName, nil
	}

	// check URL
	modelName = getModelFromURL(c.DefaultQuery("model", ""), container)
	if !utils.IsStringEmpty(modelName) {
		return modelName, nil
	}

	// check default model
	modelName = getDefaultModelName(container)
	if !utils.IsStringEmpty(modelName) {
		return modelName, nil
	}

	// model is empty
	return "", fmt.Errorf("model %s not available in publicationPoint %s and campaign %s", modelName, container.PublicationPoint, container.Campaign)
}

func getModelFromURL(modelName string, container *models.Container) string {
	if modelName != "" {
		// check if there are models available in the container
		if len(container.Models) > 0 && utils.StringInSlice(modelName, container.Models) {
			return modelName
		}
	}
	return ""
}

func getModelFromTucson(c *gin.Context, publicationPoint, campaign string) string {
	if tc, exists := c.Get("TucsonClient"); exists {
		// get model name from tucson
		if mn, err := tc.(*tucson.Client).GetModel(publicationPoint, campaign); mn != "" {
			return mn
		} else if err != nil {
			zerolog.Error().Msg(err.Error())
		}
	}
	return ""
}

func getDefaultModelName(container *models.Container) string {
	if len(container.Models) > 0 {
		// TODO: define potential default model in the future
		// return the first for now
		return container.Models[0]
	}
	return ""
}

func validateRecommendQueryParameters(rr *RecommendRequest, publicationPoint, campaign, signalID string) error {
	var mp []string

	// TODO: improve this in somehow
	if publicationPoint == "" {
		mp = append(mp, "publicationPoint")
	}

	if campaign == "" {
		mp = append(mp, "campaign")
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
	rr.SignalID = signalID

	return nil
}

// ConvertSingleEntry This function converts the Bins in the appropriate type for consistency
// The objects coming from Aerospike that have type []interface{}.
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
