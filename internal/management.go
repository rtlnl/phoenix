package internal

import (
	"errors"
	"net/http"
	"strings"

	"github.com/rtlnl/data-personalization-api/models"

	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/validator.v9"

	"github.com/rtlnl/data-personalization-api/pkg/db"
	"github.com/rtlnl/data-personalization-api/utils"
)

// ManagementModelRequest is the object that represents the payload of the request for the /management/model endpoints
type ManagementModelRequest struct {
	PublicationPoint string   `json:"publicationPoint" description:"publication point name for the model" binding:"required"`
	Campaign         string   `json:"campaign" description:"campaign name for the model" binding:"required"`
	SignalOrder      []string `json:"signalOrder" description:"list of ordered signals" binding:"required"`
	Concatenator     string   `json:"concatenator" description:"character used as concatenator for SignalOrder {'|', '#', '_', '-'}"`
}

// ManagementModelResponse is the object that represents the payload of the response for the /management/model endpoints
type ManagementModelResponse struct {
	Model   *models.Model `json:"model" description:"model object that is being returned to the client"`
	Message string        `json:"message" description:"summary of the action just taken"`
}

var (
	concatenatorList = []string{"|", "#", "_", "-"}
	validate         *validator.Validate
)

// ManagementModelRequestStructureValidation validates structure and content
func ManagementModelRequestStructureValidation(sl validator.StructLevel) {
	request := sl.Current().Interface().(ManagementModelRequest)

	// Enforces the need of a separator when more than one element
	if len(request.SignalOrder) > 1 && !utils.StringInSlice(request.Concatenator, concatenatorList) {
		sl.ReportError(request.Concatenator, "concatenator", "", "wrongConcatenator", "")
	} else if len(request.SignalOrder) == 1 && len(request.Concatenator) > 0 {
		sl.ReportError(request.Concatenator, "concatenator", "", "noConcatenatorNeeded", "")
	}
}

// ManagementModelRequestValidation validates that there are no errors in the ManagementModelRequest interface
func ManagementModelRequestValidation(request *ManagementModelRequest) error {
	validate = validator.New()
	validate.RegisterStructValidation(ManagementModelRequestStructureValidation, ManagementModelRequest{})

	return validate.Struct(request)
}

// ValidationConcatenationErrorMsg customized error message for the validation
func ValidationConcatenationErrorMsg(err error) error {
	switch {
	case strings.Contains(err.Error(), "wrongConcatenator"):
		err = errors.New("for two or more signalOrder, a concatenator character from this list is mandatory: [" + strings.Join(concatenatorList, ", ") + "]")
	case strings.Contains(err.Error(), "noConcatenatorNeeded"):
		err = errors.New("for one signalOrder no concatenator character is required")
	}
	return err
}

// GetModel returns the model's information from the given parameters in input
func GetModel(c *gin.Context) {
	ac := c.MustGet("AerospikeClient").(*db.AerospikeClient)

	// read from params in url
	pp := c.Query("publicationPoint")
	cm := c.Query("campaign")

	// if either is empty then
	if pp == "" || cm == "" {
		utils.ResponseError(c, http.StatusBadRequest, errors.New("missing parameters in url for searching the model"))
		return
	}

	// fetch model
	m, err := models.GetExistingModel(pp, cm, ac)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}

	utils.Response(c, http.StatusOK, m)
}

// CreateModel create a new model in the database where to upload the data
func CreateModel(c *gin.Context) {
	ac := c.MustGet("AerospikeClient").(*db.AerospikeClient)

	var mm ManagementModelRequest
	if err := c.BindJSON(&mm); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, ValidationConcatenationErrorMsg(err))
		return
	}

	if err := ManagementModelRequestValidation(&mm); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, ValidationConcatenationErrorMsg(err))
		return
	}

	m, err := models.NewModel(mm.PublicationPoint, mm.Campaign, mm.Concatenator, mm.SignalOrder, ac)
	if err != nil {
		utils.ResponseError(c, http.StatusUnprocessableEntity, err)
		return
	}

	utils.Response(c, http.StatusCreated, &ManagementModelResponse{
		Model:   m,
		Message: "model created",
	})
}

// PublishModel set a model to be the one to be used by the clients
func PublishModel(c *gin.Context) {
	ac := c.MustGet("AerospikeClient").(*db.AerospikeClient)

	var mm ManagementModelRequest
	if err := c.BindJSON(&mm); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	if err := ManagementModelRequestValidation(&mm); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, ValidationConcatenationErrorMsg(err))
		return
	}

	m, err := models.GetExistingModel(mm.PublicationPoint, mm.Campaign, ac)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}

	if err := m.PublishModel(ac); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	utils.Response(c, http.StatusOK, &ManagementModelResponse{
		Model:   m,
		Message: "model published",
	})
}

// StageModel set a model to be the one to be used by the internal systems
func StageModel(c *gin.Context) {
	ac := c.MustGet("AerospikeClient").(*db.AerospikeClient)

	var mm ManagementModelRequest
	if err := c.BindJSON(&mm); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	if err := ManagementModelRequestValidation(&mm); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, ValidationConcatenationErrorMsg(err))
		return
	}

	m, err := models.GetExistingModel(mm.PublicationPoint, mm.Campaign, ac)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}

	if err := m.StageModel(ac); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	utils.Response(c, http.StatusOK, &ManagementModelResponse{
		Model:   m,
		Message: "model staged",
	})
}

// EmptyModel truncate the content of a model but leave the model in the database
func EmptyModel(c *gin.Context) {
	ac := c.MustGet("AerospikeClient").(*db.AerospikeClient)

	var mm ManagementModelRequest
	if err := c.BindJSON(&mm); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	if err := ManagementModelRequestValidation(&mm); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, ValidationConcatenationErrorMsg(err))
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

	utils.Response(c, http.StatusOK, &ManagementModelResponse{
		Model:   m,
		Message: "model empty",
	})
}
