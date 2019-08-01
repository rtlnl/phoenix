package internal

import (
	"net/http"

	"github.com/rtlnl/data-personalization-api/pkg/db"
	"github.com/rtlnl/data-personalization-api/utils"

	"github.com/gin-gonic/gin"
)

// StreamingRequest is the object that represents the payload for the request in the streaming endpoints
type StreamingRequest struct {
	Signal           string   `json:"signal"`
	PublicationPoint string   `json:"publicationPoint"`
	Campaign         string   `json:"campaign"`
	Recommendations  []string `json:"recommendations"`
}

// StreamingResponse is the object that represents the payload for the response in the streaming endpoints
type StreamingResponse struct {
	Summary string `json:"summary"`
}

// CreateStreaming creates a new record
func CreateStreaming(c *gin.Context) {
	_ = c.MustGet("AerospikeClient").(*db.AerospikeClient)
	_ = c.MustGet("S3Client").(*db.S3Client)

	var sr StreamingRequest
	if err := c.BindJSON(&sr); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	utils.Response(c, http.StatusCreated, "import succeeded")
}

// UpdateStreaming updates a single record
func UpdateStreaming(c *gin.Context) {
	_ = c.MustGet("AerospikeClient").(*db.AerospikeClient)
	_ = c.MustGet("S3Client").(*db.S3Client)

	var sr StreamingRequest
	if err := c.BindJSON(&sr); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	utils.Response(c, http.StatusCreated, "import succeeded")
}

// DeleteStreaming deletes a single record
func DeleteStreaming(c *gin.Context) {
	_ = c.MustGet("AerospikeClient").(*db.AerospikeClient)
	_ = c.MustGet("S3Client").(*db.S3Client)

	var sr StreamingRequest
	if err := c.BindJSON(&sr); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	utils.Response(c, http.StatusCreated, "import succeeded")
}

// BatchRequest is the object that represents the payload of the request for the batch endpoints
type BatchRequest struct {
	Action           string      `json:"action"`
	PublicationPoint string      `json:"publicationPoint"`
	Campaign         string      `json:"campaign"`
	Data             []BatchData `json:"data" description:"used for uploading some information directly from the request"`
	DataLocation     string      `json:"dataLocation" description:"used for specifying where the data lives in S3"`
}

// BatchData is the object representing the content of the data parameter in the batch request
type BatchData map[string][]string

// BatchResponse is the object that represents the payload of the response for the batch endpoints
type BatchResponse struct {
	Summary string `json:"summary"`
}

// Batch will upload in batch a set to the database
func Batch(c *gin.Context) {
	_ = c.MustGet("AerospikeClient").(*db.AerospikeClient)

	var br BatchRequest
	if err := c.BindJSON(&br); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	utils.Response(c, http.StatusCreated, "import successed")
}
