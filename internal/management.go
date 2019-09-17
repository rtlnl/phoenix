package internal

import (
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/rtlnl/data-personalization-api/models"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gopkg.in/go-playground/validator.v8"

	"github.com/rtlnl/data-personalization-api/pkg/db"
	"github.com/rtlnl/data-personalization-api/utils"
)

// ManagementModelRequest is the object that represents the payload of the request for the /management/model endpoints
type ManagementModelRequest struct {
	PublicationPoint string   `json:"publicationPoint" description:"publication point name for the model" binding:"required"`
	Campaign         string   `json:"campaign" description:"campaign name for the model" binding:"required"`
	SignalOrder      []string `json:"signalOrder" description:"list of ordered signals" binding:"required"`
	Concatenator     string   `json:"concatenator" binding:"required,contatenatorvalidator" valid_value:"[|,#,_,-]" description:"concatenator character for signals"`
}

// ManagementModelResponse is the object that represents the payload of the response for the /management/model endpoints
type ManagementModelResponse struct {
	Message string `json:"message" description:"summary of the action just taken"`
}

var ConcatenatorList = []string{"|", "#", "_", "-"}

func ContatenatorValidator(
	v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value,
	field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string,
) bool {
	if input, ok := field.Interface().(string); ok {
		if StringInSlice(input, ConcatenatorList) {
			return false
		}
	}
	return true
}

func StringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
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

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("contatenatorvalidator", ContatenatorValidator)
	}

	var mm ManagementModelRequest
	if err := c.BindJSON(&mm); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	_, err := models.NewModel(mm.PublicationPoint, mm.Campaign, strings.Join(mm.SignalOrder, mm.Concatenator), ac)
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
