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
	Concatenator     string   `json:"concatenator" description:"character used as concatenator for SignalOrder {"|", "#", "_", "-"}"`
}

// ManagementModelDatabase is the object that represents the payload of the database schema
type ManagementModelDatabase struct {
	PublicationPoint string `json:"publicationPoint" description:"publication point name for the model" binding:"required"`
	Campaign         string `json:"campaign" description:"campaign name for the model" binding:"required"`
	Signal           string `json:"signal" description:"signals" binding:"required"`
}

// ManagementModelResponse is the object that represents the payload of the response for the /management/model endpoints
type ManagementModelResponse struct {
	Message string `json:"message" description:"summary of the action just taken"`
}

var (
	concatenatorList = []string{"|", "#", "_", "-"}
	validate         *validator.Validate
)

// Checks if a string was found in a slice
func StringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

// Validate structure and content
func ManagementModelRequestStructureValidation(structLevel validator.StructLevel) {
	request := structLevel.Current().Interface().(ManagementModelRequest)

	// Enforces the need of a separator when more than one element
	if len(request.SignalOrder) > 1 && !StringInSlice(request.Concatenator, concatenatorList) {
		structLevel.ReportError(request.Concatenator, "concatenator", "", "wrongConcatenator", "")
	} else if len(request.SignalOrder) == 1 && len(request.Concatenator) > 0 {
		structLevel.ReportError(request.Concatenator, "concatenator", "", "noConcatenatorNeeded", "")
	}
}

// Validate that there are no errors in the ManagementModelRequest interface
func ManagementModelRequestValidation(request *ManagementModelRequest) error {
	validate = validator.New()
	validate.RegisterStructValidation(ManagementModelRequestStructureValidation, ManagementModelRequest{})

	err := validate.Struct(request)
	if err != nil {
		return err
	}
	return nil
}

// Fills up the database schema
func GetManagementModelAttributes(request *ManagementModelRequest) ManagementModelDatabase {
	var result ManagementModelDatabase

	result.PublicationPoint = request.PublicationPoint
	result.Campaign = request.Campaign

	// If more than one member in the slice, join them using the concatenator
	if len(request.SignalOrder) == 1 {
		result.Signal = string(request.SignalOrder[0])
	} else if len(request.SignalOrder) > 1 {
		result.Signal = strings.Join(request.SignalOrder, request.Concatenator)
	}

	return result
}

// Customized error message for the validation
func ValidationErrorMessage(err error) error {
	if strings.Contains(err.Error(), "wrongConcatenator") {
		err = errors.New("for two or more signalOrder, a concatenator character from this list is mandatory: [" + strings.Join(concatenatorList, ", ") + "]")
	} else if strings.Contains(err.Error(), "noConcatenatorNeeded") {
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
		utils.ResponseError(c, http.StatusBadRequest, ValidationErrorMessage(err))
		return
	}

	if err := ManagementModelRequestValidation(&mm); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, ValidationErrorMessage(err))
		return
	}

	mmdb := GetManagementModelAttributes(&mm)

	_, err := models.NewModel(mmdb.PublicationPoint, mmdb.Campaign, mmdb.Signal, ac)
	if err != nil {
		utils.ResponseError(c, http.StatusUnprocessableEntity, err)
		return
	}

	utils.Response(c, http.StatusCreated, &ManagementModelResponse{
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
		utils.ResponseError(c, http.StatusBadRequest, ValidationErrorMessage(err))
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
		utils.ResponseError(c, http.StatusBadRequest, ValidationErrorMessage(err))
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
		utils.ResponseError(c, http.StatusBadRequest, ValidationErrorMessage(err))
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
		Message: "model empty",
	})
}
