package public

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	zerolog "github.com/rs/zerolog/log"

	"github.com/rtlnl/phoenix/models"
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
// and efficiently improve the garbage collection operations. rr is short for recommend-request
var rrPool = sync.Pool{
	New: func() interface{} { return new(RecommendRequest) },
}

// Recommend will take care of fetching the personalized content for a specific user
func Recommend(c *gin.Context) {
	dbc := c.MustGet("DB").(db.DB)
	lt := c.MustGet("RecommendationLog").(logs.RecommendationLog)

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

	// get container from DB
	container, err := models.GetContainer(rr.PublicationPoint, rr.Campaign, dbc)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}

	// get model name either from Tucson or URL
	modelName, err := getModelName(c, container)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}

	// model exists
	m, err := models.GetModel(modelName, dbc)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}

	// validate signal
	if !m.CorrectSignalFormat(rr.SignalID) {
		utils.ResponseError(c, http.StatusBadRequest, errors.New("signal is not formatted correctly"))
		return
	}

	// get the recommended values
	r, err := dbc.GetOne(modelName, rr.SignalID)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}

	// convert single entry from string to []models.ItemScore
	itemsScore, err := models.DeserializeItemScoreArray(r)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, fmt.Errorf("could not deserialize object. error: %s", err.Error()))
		return
	}

	// write logs in a separate thread for not blocking the server
	go func() {
		err := lt.Write(logs.RowLog{
			PublicationPoint: rr.PublicationPoint,
			Campaign:         rr.Campaign,
			SignalID:         rr.SignalID,
			ItemScores:       itemsScore,
		})
		// log error if it fails the logging
		if err != nil {
			zerolog.Error().Msg(err.Error())
		}
	}()

	utils.Response(c, http.StatusOK, &RecommendResponse{
		ModelName:       modelName,
		Recommendations: itemsScore,
	})
}

func getModelName(c *gin.Context, container models.Container) (string, error) {
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

func getModelFromURL(modelName string, container models.Container) string {
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

func getDefaultModelName(container models.Container) string {
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
