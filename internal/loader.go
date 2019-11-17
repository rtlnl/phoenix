package internal

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"

	"github.com/rtlnl/phoenix/models"
	"github.com/rtlnl/phoenix/pkg/db"
	"github.com/rtlnl/phoenix/utils"
)

// used to fast unmarshal json strings
var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	// name of the table for storing all the batchIDs
	tableBulkStatus = "bulkStatus"
	// name of the error tables for storing all the errors of a specific batch
	tableBulkErrors = "bulkErrors"
	// numberErrors
	maxErrorLines = 50
)

// StreamingRequest is the object that represents the payload for the request in the streaming endpoints
type StreamingRequest struct {
	Signal          string             `json:"signal" binding:"required"`
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
	if m.RequireSignalFormat() && !m.CorrectSignalFormat(sr.Signal) {
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
	if err := dbc.AddOne(sr.ModelName, sr.Signal, ser); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	utils.Response(c, http.StatusCreated, &StreamingResponse{
		Message: fmt.Sprintf("signal %s created", sr.Signal),
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
	if m.RequireSignalFormat() && !m.CorrectSignalFormat(sr.Signal) {
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
	if err := dbc.AddOne(sr.ModelName, sr.Signal, ser); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return

	}
	utils.Response(c, http.StatusOK, &StreamingResponse{
		Message: fmt.Sprintf("signal %s updated", sr.Signal),
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
	if err := dbc.DeleteOne(sr.ModelName, sr.Signal); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	utils.Response(c, http.StatusOK, &StreamingResponse{
		Message: fmt.Sprintf("signal %s deleted", sr.Signal),
	})
}

// BatchRequest is the object that represents the payload of the request for the batch endpoints
// Conditions: Data takes precedence in case also DataLocation is specified
type BatchRequest struct {
	ModelName    string      `json:"modelName" binding:"required"`
	Data         []BatchData `json:"data" description:"used for uploading some information directly from the request"`
	DataLocation string      `json:"dataLocation" description:"used for specifying where the data lives in S3"`
}

// BatchData is the object representing the content of the data parameter in the batch request
type BatchData map[string][]models.ItemScore

// BatchResponse is the object that represents the payload of the response for the batch endpoints
type BatchResponse struct {
	NumberOfLines string `json:"numberoflines" description:"total count of lines"`
	ErrorRecords  DataUploadedError
}

// BatchStatusResponseError is the response paylod when the batch upload failed
type BatchStatusResponseError struct {
	Status              string             `json:"status" description:"define the status of the bulk upload when importing data from a file"`
	NumberOfLinesFailed string             `json:"numberoflinesfailed" description:"total count of failed lines"`
	Line                []models.LineError `json:"line" description:"shows the line error and the reason, i.e. {'100', 'reason': 'validation message'}"`
}

// BatchBulkResponse is the object that represents the payload of the response when uploading from S3
type BatchBulkResponse struct {
	BatchID string `json:"batchId"`
}

// DataUploadedError is the response payload when the batch upload failed
type DataUploadedError struct {
	NumberOfLinesFailed string             `json:"numberoflinesfailed" description:"total count of lines that were not uploaded"`
	Errors              []models.LineError `json:"error" description:"errors found"`
}

// Batch will upload in batch a set to the database
func Batch(c *gin.Context) {
	dbc := c.MustGet("DB").(db.DB)
	sess := c.MustGet("AWSSession").(*session.Session)

	var br BatchRequest
	if err := c.BindJSON(&br); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	// retrieve the model
	m, err := models.GetModel(br.ModelName, dbc)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}

	// upload data from request itself
	bo := NewBatchOperator(dbc, m)
	if len(br.Data) > 0 && br.Data != nil {
		ln, due, err := bo.UploadDataDirectly(br.Data)
		if err != nil {
			utils.ResponseError(c, http.StatusInternalServerError, err)
			return
		}
		utils.Response(c, http.StatusCreated, &BatchResponse{NumberOfLines: ln, ErrorRecords: due})
		return
	}

	// truncate eventual old data
	if err := dbc.DropTable(br.ModelName); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	// upload data from S3 file
	bucket, key := utils.StripS3URL(br.DataLocation)
	s := db.NewS3Client(&db.S3Bucket{Bucket: bucket, ACL: ""}, sess)

	// check if file exists
	if s.ExistsObject(key) == false {
		utils.ResponseError(c, http.StatusBadRequest, fmt.Errorf("key %s not founds in S3", br.DataLocation))
		return
	}
	// download the file
	f, err := s.GetObject(key)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	// generate batchID
	batchID := uuid.New().String()

	// seprate thread
	go bo.UploadDataFromFile(f, batchID)

	utils.Response(c, http.StatusCreated, &BatchBulkResponse{BatchID: batchID})
}

// BatchStatusResponse is the response payload for getting the status of the bulk upload from S3
type BatchStatusResponse struct {
	Status string             `json:"status"`
	Errors []models.LineError `json:"errors"`
}

// BatchStatus returns the current status of the batch upload
func BatchStatus(c *gin.Context) {
	dbc := c.MustGet("DB").(db.DB)
	batchID := c.Param("id")

	// get the status of the batch
	status, err := dbc.GetOne(tableBulkStatus, batchID)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, fmt.Errorf("batch job with ID %s not found", batchID))
		return
	}

	switch status {
	case BulkPartialUpload:
		// get from table errors
		ser, err := dbc.GetOne(tableBulkErrors, batchID)
		if err != nil {
			utils.ResponseError(c, http.StatusInternalServerError, err)
			return
		}
		// deserialize object
		errs, err := models.DeserializeLineErrorArray(ser)
		if err != nil {
			utils.ResponseError(c, http.StatusInternalServerError, err)
			return
		}
		utils.Response(c, http.StatusOK, &BatchStatusResponse{Status: status, Errors: errs})
		break
	default:
		utils.Response(c, http.StatusOK, &BatchStatusResponse{Status: status})
	}
}
