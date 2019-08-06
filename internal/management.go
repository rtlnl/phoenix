package internal

import (
	"fmt"
	"net/http"

	"github.com/rtlnl/data-personalization-api/models"

	"github.com/gin-gonic/gin"
	"github.com/rtlnl/data-personalization-api/pkg/db"
	"github.com/rtlnl/data-personalization-api/utils"
)

// ManagementModelRequest is the object that represents the payload of the request for the /management/model endpoints
type ManagementModelRequest struct {
	PublicationPoint string   `json:"publicationPoint" description:"publication point name for the model"`
	Campaign         string   `json:"campaign" description:"campaign name for the model"`
	Type             string   `json:"type" description:"type of upload for the model"`
	Signals          []string `json:"signals" description:"list of signals name"`
	SignalOrder      string   `json:"signalOrder" description:"signal order definition"`
}

// ManagementModelResponse is the object that represents the payload of the response for the /management/model endpoints
type ManagementModelResponse struct {
	Summary string `json:"summary" description:"summary of the action just taken"`
}

// CreateModel create a new model in the database where to upload the data
func CreateModel(c *gin.Context) {
	ac := c.MustGet("AerospikeClient").(*db.AerospikeClient)

	var mm ManagementModelRequest
	if err := c.BindJSON(&mm); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	_, err := models.NewModel(mm.PublicationPoint, mm.Campaign, mm.SignalOrder, ac)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	utils.Response(c, http.StatusCreated, &ManagementModelResponse{
		Summary: fmt.Sprintf("create %s %s %s", mm.PublicationPoint, mm.Campaign, mm.SignalOrder),
	})
}

// DeleteModel deletes a model in the database and all its related data
func DeleteModel(c *gin.Context) {
	ac := c.MustGet("AerospikeClient").(*db.AerospikeClient)

	var mm ManagementModelRequest
	if err := c.BindJSON(&mm); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	m, err := models.GetExistingModel(mm.PublicationPoint, mm.Campaign, ac)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}

	// delete model from database
	if err := m.DeleteModel(ac); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	utils.Response(c, http.StatusCreated, "model deletion succeeded")
}

// PublishModel set a model to be the one to be used by the clients
func PublishModel(c *gin.Context) {
	ac := c.MustGet("AerospikeClient").(*db.AerospikeClient)

	var mm ManagementModelRequest
	if err := c.BindJSON(&mm); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	m, err := models.GetExistingModel(mm.PublicationPoint, mm.Campaign, ac)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}

	if err := m.PublishModel(ac); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	utils.Response(c, http.StatusCreated, "model publication succeeded")
}

// EmptyModel truncate the content of a model but leave the model in the database
func EmptyModel(c *gin.Context) {
	ac := c.MustGet("AerospikeClient").(*db.AerospikeClient)

	var mm ManagementModelRequest
	if err := c.BindJSON(&mm); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	m, err := models.GetExistingModel(mm.PublicationPoint, mm.Campaign, ac)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}

	// empty model from database
	if err := m.DeleteModel(ac); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	utils.Response(c, http.StatusCreated, "model empty succeeded")
}
