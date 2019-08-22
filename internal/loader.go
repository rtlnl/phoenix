package internal

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-redis/redis"
	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/rtlnl/data-personalization-api/models"

	"github.com/rtlnl/data-personalization-api/pkg/db"
	"github.com/rtlnl/data-personalization-api/utils"

	"github.com/gin-gonic/gin"
)

const (
	// The current CSV file we read contains only 2 columns: signal;items
	columnsDataFile = 2
	// csv delimiter character
	csvDelimiter = ";"
	// record delimiter that separates each recommended item
	recordDelimiter = ","
	// name of the key of the bins containing the recommended items
	// Max 15 characters due to Aerospike limitation
	binKey = "items"
)

// StreamingRequest is the object that represents the payload for the request in the streaming endpoints
type StreamingRequest struct {
	Signal           string   `json:"signal" binding:"required"`
	PublicationPoint string   `json:"publicationPoint" binding:"required"`
	Campaign         string   `json:"campaign" binding:"required"`
	Recommendations  []string `json:"recommendations" binding:"required"`
}

// StreamingResponse is the object that represents the payload for the response in the streaming endpoints
type StreamingResponse struct {
	Message string `json:"message"`
}

// CreateStreaming creates a new record in the selected campaign
func CreateStreaming(c *gin.Context) {
	ac := c.MustGet("AerospikeClient").(*db.AerospikeClient)

	var sr StreamingRequest
	if err := c.BindJSON(&sr); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	m, err := models.GetExistingModel(sr.PublicationPoint, sr.Campaign, ac)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}

	// cannot add data to published models
	if m.Stage == models.PUBLISHED {
		utils.ResponseError(c, http.StatusBadRequest, errors.New("you cannot add data on already published models. stage it first"))
		return
	}

	sn := m.ComposeSetName()
	if err := ac.AddOne(sn, sr.Signal, binKey, sr.Recommendations); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	utils.Response(c, http.StatusCreated, &StreamingResponse{
		Message: fmt.Sprintf("signal %s created", sr.Signal),
	})
}

// UpdateStreaming updates a single record in the selected campaign
func UpdateStreaming(c *gin.Context) {
	ac := c.MustGet("AerospikeClient").(*db.AerospikeClient)

	var sr StreamingRequest
	if err := c.BindJSON(&sr); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	m, err := models.GetExistingModel(sr.PublicationPoint, sr.Campaign, ac)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}

	// cannot update data to published models
	if m.Stage == models.PUBLISHED {
		utils.ResponseError(c, http.StatusBadRequest, errors.New("you cannot update data on already published models. stage it first"))
		return
	}

	sn := m.ComposeSetName()

	// The AddOne method does an UPSERT
	if err := ac.AddOne(sn, sr.Signal, binKey, sr.Recommendations); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return

	}
	utils.Response(c, http.StatusOK, &StreamingResponse{
		Message: fmt.Sprintf("signal %s updated", sr.Signal),
	})
}

// DeleteStreaming deletes a single record in the selected campaign
func DeleteStreaming(c *gin.Context) {
	ac := c.MustGet("AerospikeClient").(*db.AerospikeClient)

	var sr StreamingRequest
	if err := c.BindJSON(&sr); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	m, err := models.GetExistingModel(sr.PublicationPoint, sr.Campaign, ac)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, err)
		return
	}

	// cannot delete data to published models
	if m.Stage == models.PUBLISHED {
		utils.ResponseError(c, http.StatusBadRequest, errors.New("you cannot delete data on already published models. stage it first"))
		return
	}

	sn := m.ComposeSetName()
	if err := ac.DeleteOne(sn, sr.Signal); err != nil {
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
	PublicationPoint string      `json:"publicationPoint"`
	Campaign         string      `json:"campaign"`
	Data             []BatchData `json:"data" description:"used for uploading some information directly from the request"`
	DataLocation     string      `json:"dataLocation" description:"used for specifying where the data lives in S3"`
}

// BatchData is the object representing the content of the data parameter in the batch request
type BatchData map[string][]string

// BatchResponse is the object that represents the payload of the response for the batch endpoints
type BatchResponse struct {
	Message string `json:"message"`
}

// BatchBulkResponse is the object that represents the payload of the response when uploading from S3
type BatchBulkResponse struct {
	BatchID string `json:"batchId"`
}

// DataUploaded is the object that represents the result for uplaoding the data
type DataUploaded struct {
	NumberOfSignals         int `description:"total count of signals that have been uploaded"`
	NumberOfRecommendations int `description:"total count of recommendations items that have been uploaded"`
}

// Batch will upload in batch a set to the database
func Batch(c *gin.Context) {
	ac := c.MustGet("AerospikeClient").(*db.AerospikeClient)
	rc := c.MustGet("RedisClient").(*redis.Client)
	sess := c.MustGet("AWSSession").(*session.Session)

	var br BatchRequest
	if err := c.BindJSON(&br); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	// retrieve the model
	m, err := models.GetExistingModel(br.PublicationPoint, br.Campaign, ac)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, fmt.Errorf("model with publicationPoint %s and campaign %s not found", br.PublicationPoint, br.Campaign))
		return
	}

	// cannot uplaod data to published models
	if m.Stage == models.PUBLISHED {
		utils.ResponseError(c, http.StatusBadRequest, errors.New("you cannot upload data on already published models. stage it first"))
		return
	}

	// upload data from request itself
	if len(br.Data) > 0 && br.Data != nil {
		du, err := uploadDataDirectly(ac, br.Data, m)
		if du == nil && err != nil {
			utils.ResponseError(c, http.StatusInternalServerError, err)
			return
		}
		utils.Response(c, http.StatusCreated, &BatchResponse{Message: "data uploaded"})
		return
	}

	// upload data from S3 file
	bucket, key := StripS3URL(br.DataLocation)
	s := db.NewS3Client(bucket, sess)

	// check if file exists
	if s.ExistsObject(key) == false {
		utils.ResponseError(c, http.StatusBadRequest, fmt.Errorf("key %s does not exists in S3", br.DataLocation))
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
	go uploadDataFromFile(ac, rc, f, m, batchID)

	utils.Response(c, http.StatusCreated, &BatchBulkResponse{BatchID: batchID})
}

// StripS3URL returns the bucket and the key from a s3 url location
func StripS3URL(URL string) (string, string) {
	bucketTmp := strings.Replace(URL, "s3://", "", -1)

	bucket := bucketTmp[:strings.IndexByte(bucketTmp, '/')]
	key := strings.TrimPrefix(URL, fmt.Sprintf("s3://%s/", bucket))

	return bucket, key
}

func uploadDataDirectly(ac *db.AerospikeClient, bd []BatchData, m *models.Model) (*DataUploaded, error) {
	var nr int
	for _, data := range bd {
		for sig, recommendedItems := range data {
			nr += len(recommendedItems)

			// upload to Aerospike
			// TODO: verify here if it works in this way
			setName := m.ComposeSetName()
			if err := ac.AddOne(setName, sig, binKey, recommendedItems); err != nil {
				return nil, err
			}
		}
	}
	return &DataUploaded{NumberOfSignals: len(bd), NumberOfRecommendations: nr}, nil
}

// BulkStatus defines the status of the batch upload from S3
type BulkStatus string

const (
	// BulkUploading represents the uploading status
	BulkUploading = "UPLOADING"
	// BulkSucceeded represents the succeeded status
	BulkSucceeded = "SUCCEEDED"
	// BulkFailed represents the failed status
	BulkFailed = "FAILED"
)

func uploadDataFromFile(ac *db.AerospikeClient, rc *redis.Client, file *io.ReadCloser, m *models.Model, batchID string) {

	// write to Redis it's uploading
	if err := rc.Set(batchID, BulkUploading, 0).Err(); err != nil {
		// if this fails than since we cannot return the request to the user
		// we need to restart the application
		log.Panicln(err)
	}

	records := 0
	rd := bufio.NewReader(*file)

	var nr int
	for {
		line, err := rd.ReadString('\n')
		if err == io.EOF {
			break
		}

		// string new-line character
		l := strings.TrimSuffix(line, "\n")

		// skip header in csv
		// TODO: do we need to skip the header??
		if records <= 0 {
			records++
			continue
		}

		record := strings.Split(l, csvDelimiter)
		if len(record) != columnsDataFile {
			continue
		}

		sig := record[0]
		recommendedItems := strings.Split(record[1], recordDelimiter)

		// count number of recommendations
		nr += len(recommendedItems)

		// upload to Aerospike
		setName := m.ComposeSetName()
		if err := ac.AddOne(setName, sig, binKey, recommendedItems); err != nil {
			// write to Redis it failed
			if err := rc.Set(batchID, BulkFailed, 0).Err(); err != nil {
				// if this fails than since we cannot return the request to the user
				// we need to restart the application
				log.Panicln(err)
			}
			return
		}
		records++
	}

	// write to Redis it succeeded
	if err := rc.Set(batchID, BulkSucceeded, 0).Err(); err != nil {
		// if this fails than since we cannot return the request to the user
		// we need to restart the application
		log.Panicln(err)
	}
}

// BatchStatusResponse is the response payload for getting the status of the bulk upload from S3
type BatchStatusResponse struct {
	Status string `json:"status"`
}

// BatchStatus returns the current status of the batch upload
func BatchStatus(c *gin.Context) {
	rc := c.MustGet("RedisClient").(*redis.Client)

	batchID := c.Param("id")

	s, err := rc.Get(batchID).Result()
	if err == redis.Nil {
		utils.ResponseError(c, http.StatusNotFound, fmt.Errorf("batch job with ID %s not found", batchID))
		return
	} else if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}
	utils.Response(c, http.StatusOK, &BatchStatusResponse{Status: s})
}
