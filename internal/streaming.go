package internal

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"

	"github.com/rtlnl/phoenix/models"
	"github.com/rtlnl/phoenix/pkg/db"
	"github.com/rtlnl/phoenix/utils"
)

// used to fast unmarshal json strings
var json = jsoniter.ConfigCompatibleWithStandardLibrary

// StreamingRequest is the object that represents the payload for the request in the streaming endpoints
type StreamingRequest struct {
	SignalID        string             `json:"signalId" binding:"required"`
	ModelName       string             `json:"modelName" binding:"required"`
	Recommendations []models.ItemScore `json:"recommendations" binding:"required"`
}

// StreamingResponse is the object that represents the payload for the response in the streaming endpoints
type StreamingResponse struct {
	Message string `json:"message"`
}

// CreateStreaming creates a new record in the selected campaign
func CreateStreaming(c *gin.Context) {
	dbc := c.MustGet("DB").(db.DB)

	var sr StreamingRequest
	if err := c.BindJSON(&sr); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}
	// get the model
	m, err := models.GetModel(sr.ModelName, dbc)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}
	// validate input
	if m.RequireSignalFormat() && !m.CorrectSignalFormat(sr.SignalID) {
		utils.ResponseError(c, http.StatusBadRequest, fmt.Errorf("the expected signal format must be %s", strings.Join(m.SignalOrder, m.Concatenator)))
		return
	}
	// serialize recommendations
	ser, err := utils.SerializeObject(sr.Recommendations)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}
	// store data in the database
	if err := dbc.AddOne(sr.ModelName, sr.SignalID, ser); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	utils.Response(c, http.StatusCreated, &StreamingResponse{
		Message: fmt.Sprintf("signal %s created", sr.SignalID),
	})
}

// UpdateStreaming updates a single record in the selected campaign
func UpdateStreaming(c *gin.Context) {
	dbc := c.MustGet("DB").(db.DB)

	var sr StreamingRequest
	if err := c.BindJSON(&sr); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}
	// get model
	m, err := models.GetModel(sr.ModelName, dbc)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}
	// validate input
	if m.RequireSignalFormat() && !m.CorrectSignalFormat(sr.SignalID) {
		utils.ResponseError(c, http.StatusBadRequest, fmt.Errorf("the expected signal format must be %s", strings.Join(m.SignalOrder, m.Concatenator)))
		return
	}
	// serialize recommendations
	ser, err := utils.SerializeObject(sr.Recommendations)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}
	// The AddOne method does an UPSERT
	if err := dbc.AddOne(sr.ModelName, sr.SignalID, ser); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return

	}
	utils.Response(c, http.StatusOK, &StreamingResponse{
		Message: fmt.Sprintf("signal %s updated", sr.SignalID),
	})
}

// DeleteStreaming deletes a single record in the selected campaign
func DeleteStreaming(c *gin.Context) {
	dbc := c.MustGet("DB").(db.DB)

	var sr StreamingRequest
	if err := c.BindJSON(&sr); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}
	// get model
	exists := models.ModelExists(sr.ModelName, dbc)
	if !exists {
		utils.ResponseError(c, http.StatusNotFound, fmt.Errorf("model %s not found", sr.ModelName))
		return
	}
	// delete record
	if err := dbc.DeleteOne(sr.ModelName, sr.SignalID); err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}

	utils.Response(c, http.StatusOK, &StreamingResponse{
		Message: fmt.Sprintf("signal %s deleted", sr.SignalID),
	})
}

// RecommendationRequest is the object that represents the payload for the request
//in the recommendation streaming endpoint
type RecommendationRequest struct {
	SignalID       string           `json:"signalId" binding:"required"`
	ModelName      string           `json:"modelName" binding:"required"`
	Recommendation models.ItemScore `json:"recommendation" binding:"required"`
}

// DeleteRecommendation handles a deletion for a single recommendation item in the list
// given a signalId and modelName
func DeleteRecommendation(c *gin.Context) {
	dbc := c.MustGet("DB").(db.DB)

	var lr RecommendationRequest
	if err := c.BindJSON(&lr); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	// get the model
	m, err := models.GetModel(lr.ModelName, dbc)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}

	// validate input
	if m.RequireSignalFormat() && !m.CorrectSignalFormat(lr.SignalID) {
		utils.ResponseError(c, http.StatusBadRequest, fmt.Errorf("the expected signal format must be %s", strings.Join(m.SignalOrder, m.Concatenator)))
		return
	}

	// get the recommended values
	rec, err := dbc.GetOne(lr.ModelName, lr.SignalID)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}

	// convert the escaped json string to itemscore object
	var items []models.ItemScore
	if err := json.Unmarshal([]byte(rec), &items); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	// remove the item from the recommendation list
	// only do so if it actually exists
	var valid bool
	items, valid = removeItem(lr.Recommendation, items)
	if !valid {
		utils.ResponseError(c, http.StatusBadRequest, errors.New("recommendation does not exist"))
		return
	}

	// serialize recommendations
	ser, err := utils.SerializeObject(items)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	// UPSERT the new recommendation list to the DB
	if err := dbc.AddOne(lr.ModelName, lr.SignalID, ser); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	log.Info().Str("DELETE", fmt.Sprintf("SignalId %s", lr.SignalID)).Str("MODEL", fmt.Sprintf("name %s", lr.ModelName))

	utils.Response(c, http.StatusCreated, &StreamingResponse{
		Message: fmt.Sprintf("Handled recommended item deletion for SignalId %s", lr.SignalID),
	})
}

// Removes a given itemscore item from the itemscore array
// and returns an array with the item removed and a boolean if the removal
// was succesful. True if the item was found and removed, false if it was not found.
func removeItem(toRemove models.ItemScore, items []models.ItemScore) ([]models.ItemScore, bool) {
	found := false

	// If items is not valid, return immediately
	if len(items) == 0 {
		return items, found
	}

	tmp := items[:0]
	for _, item := range items {
		if !(toRemove["item"] == item["item"]) {
			tmp = append(tmp, item)
		} else {
			found = true
		}
	}
	return tmp, found
}
