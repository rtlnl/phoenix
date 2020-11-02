package internal

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/rtlnl/phoenix/models"
	"github.com/rtlnl/phoenix/pkg/batch"
	"github.com/rtlnl/phoenix/pkg/db"
	"github.com/rtlnl/phoenix/utils"
	"github.com/rtlnl/phoenix/worker"
)

// BatchRequest is the object that represents the payload of the request for the batch endpoints
// Conditions: Data takes precedence in case also DataLocation is specified
type BatchRequest struct {
	ModelName    string       `json:"modelName" binding:"required"`
	Data         []batch.Data `json:"data" description:"used for uploading some information directly from the request"`
	DataLocation string       `json:"dataLocation" description:"used for specifying where the data lives in S3"`
}

// BatchResponse is the object that represents the payload of the response for the batch endpoints
type BatchResponse struct {
	NumberOfLines string `json:"numberoflines" description:"total count of lines"`
	ErrorRecords  batch.DataUploadedError
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

// Batch will upload in batch a set to the database
func Batch(c *gin.Context) {
	dbc := c.MustGet("DB").(db.DB)
	wrk := c.MustGet("Worker").(*worker.Worker)

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
	bo := batch.NewOperator(dbc, m)
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
	log.Info().Str("DELETE", fmt.Sprintf("table %s", br.ModelName))
	// upload data from S3 file
	bucket, key := utils.StripS3URL(br.DataLocation)
	// generate batchID
	batchID := uuid.New().String()
	// get from the ENV if we need to disable SSL (used in local development)
	val := os.Getenv("S3_DISABLE_SSL")
	disableSSL, err := strconv.ParseBool(val)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}
	// write to DB that it's uploading
	if err := bo.SetStatus(batchID, batch.BulkQueued); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}
	// create task payload to send to the queue
	taskPayload := &worker.TaskPayload{
		DBURL:        os.Getenv("DB_HOST"),
		DBPassword:   os.Getenv("DB_PASSWORD"),
		AWSRegion:    os.Getenv("S3_REGION"),
		S3Endpoint:   os.Getenv("S3_ENDPOINT"),
		S3DisableSSL: disableSSL,
		S3Bucket:     bucket,
		S3Key:        key,
		ModelName:    br.ModelName,
		BatchID:      batchID,
	}
	// publish message to the queue
	if err := wrk.Publish(taskPayload); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	log.Info().Str("BATCH", fmt.Sprintf("started batchId %s", batchID)).Str("MODEL", fmt.Sprintf("name %s", br.ModelName))

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
	status, err := dbc.GetOne(batch.TableBulkStatus, batchID)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, fmt.Errorf("batch job with ID %s not found", batchID))
		return
	}

	switch status {
	case batch.BulkPartialUpload:
		// get from table errors
		ser, err := dbc.GetOne(batch.TableBulkErrors, batchID)
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
