package internal

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rtlnl/data-personalization-api/models"

	"github.com/rtlnl/data-personalization-api/pkg/db"
	"github.com/rtlnl/data-personalization-api/utils"

	"github.com/gin-gonic/gin"
)

const (
	// The current CSV file we read contains only 2 columns: signal;items
	numOfColumnsInDataFile = 2
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

// CreateStreaming creates a new record in the selected campaign
func CreateStreaming(c *gin.Context) {
	_ = c.MustGet("AerospikeClient").(*db.AerospikeClient)

	var sr StreamingRequest
	if err := c.BindJSON(&sr); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	utils.Response(c, http.StatusCreated, "import succeeded")
}

// UpdateStreaming updates a single record in the selected campaign
func UpdateStreaming(c *gin.Context) {
	_ = c.MustGet("AerospikeClient").(*db.AerospikeClient)

	var sr StreamingRequest
	if err := c.BindJSON(&sr); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	utils.Response(c, http.StatusCreated, "import succeeded")
}

// DeleteStreaming deletes a single record in the selected campaign
func DeleteStreaming(c *gin.Context) {
	_ = c.MustGet("AerospikeClient").(*db.AerospikeClient)

	var sr StreamingRequest
	if err := c.BindJSON(&sr); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	utils.Response(c, http.StatusCreated, "import succeeded")
}

// BatchRequest is the object that represents the payload of the request for the batch endpoints
// Conditions: Data takes precedence in case also DataLocation is specified
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

// DataUploaded is the object that represents the result for uplaoding the data
type DataUploaded struct {
	NumberOfSignals         int `description:"total count of signals that have been uploaded"`
	NumberOfRecommendations int `description:"total count of recommendations items that have been uploaded"`
}

// Batch will upload in batch a set to the database
func Batch(c *gin.Context) {
	ac := c.MustGet("AerospikeClient").(*db.AerospikeClient)
	s := c.MustGet("S3Client").(*db.S3Client)

	var br BatchRequest
	if err := c.BindJSON(&br); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	// retrieve the model
	m, err := models.GetExistingModel(br.PublicationPoint, br.Campaign, ac)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	var du *DataUploaded
	if len(br.Data) > 0 && br.Data != nil {
		du, err = uploadDataDirectly(ac, br.Data, m)
	} else {
		// check if file exists
		key := strings.TrimPrefix(br.DataLocation, fmt.Sprintf("s3://%s/", s.Bucket))
		if s.ExistsObject(key) == false {
			utils.ResponseError(c, http.StatusBadRequest, fmt.Errorf("key %s does not exists", br.DataLocation))
			return
		}
		// download the file
		f, err := s.GetObject(key)
		if err != nil {
			utils.ResponseError(c, http.StatusInternalServerError, err)
			return
		}
		du, err = uploadDataFromFile(ac, f, m)
	}

	if du == nil && err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	summary := fmt.Sprintf("%s %s %s %d %d", br.Action, br.PublicationPoint, br.Campaign, du.NumberOfSignals, du.NumberOfRecommendations)
	utils.Response(c, http.StatusCreated, &BatchResponse{Summary: summary})
}

func uploadDataDirectly(ac *db.AerospikeClient, bd []BatchData, m *models.Model) (*DataUploaded, error) {
	var nr int
	for _, data := range bd {
		for sig, recommendedItems := range data {
			nr += len(recommendedItems)

			// transform to be complaint with the interface
			v := make(map[string]interface{})
			v[sig] = recommendedItems

			// upload to Aerospike
			// TODO: verify here if it works in this way
			setName := m.ComposeSetName()
			if err := ac.AddOne(setName, sig, sig, v); err != nil {
				return nil, err
			}
		}
	}
	return &DataUploaded{NumberOfSignals: len(bd), NumberOfRecommendations: nr}, nil
}

func uploadDataFromFile(ac *db.AerospikeClient, file *io.ReadCloser, m *models.Model) (*DataUploaded, error) {

	records := 0
	rd := bufio.NewReader(*file)

	var nr int
	for {
		l, err := rd.ReadString('\n')
		if err == io.EOF {
			break
		}

		// skip header in csv
		if records <= 0 {
			records++
			continue
		}

		record := strings.Split(l, ";")
		if len(record) != numOfColumnsInDataFile {
			continue
		}

		sig := record[0]
		recommendedItems := strings.Split(record[1], ",")

		// count number of recommendations
		nr += len(recommendedItems)

		// transform to be complaint with the interface
		v := make(map[string]interface{})
		v[sig] = recommendedItems

		// upload to Aerospike
		// TODO: verify here if it works in this way
		setName := m.ComposeSetName()
		if err := ac.AddOne(setName, sig, sig, v); err != nil {
			return nil, err
		}
		records++
	}
	return &DataUploaded{NumberOfSignals: records, NumberOfRecommendations: nr}, nil
}
