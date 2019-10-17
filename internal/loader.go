package internal

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"

	"github.com/rtlnl/data-personalization-api/models"
	"github.com/rtlnl/data-personalization-api/pkg/db"
	"github.com/rtlnl/data-personalization-api/utils"
)

// used to fast unmarshal json strings
var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	// name of the key of the bins containing the recommended items
	// Max 15 characters due to Aerospike limitation
	binKey = "items"
	// name of the setName for storing all the batchIDs
	bulkStatusSetName = "bulkStatus"
	// name of the key to retrieve the status of the bulk upload
	statusBinKey = "status"
	// name of the key to insert the errors of failed uplaoded lines
	lineBinError = "lineError"
	// numberErrors
	maxErrorLines = 50
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

// Check if it is required to check the signal format
func requireSignalFormat(m *models.Model) bool {
	if len(m.SignalOrder) > 1 && m.Concatenator != "" {
		return true
	}
	return false
}

// Check that the signal format is correct
func correctSignalFormat(m *models.Model, s string) bool {
	res := strings.FieldsFunc(s, func(c rune) bool {
		r := []rune(m.Concatenator)
		return c == r[0]
	})
	return len(m.SignalOrder) == len(res)
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
	if m.IsPublished() {
		utils.ResponseError(c, http.StatusBadRequest, errors.New("you cannot add data on already published models. stage it first"))
		return
	}

	if requireSignalFormat(m) && !correctSignalFormat(m, sr.Signal) {
		utils.ResponseError(c, http.StatusBadRequest, errors.New("the expected signal format must be "+strings.Join(m.SignalOrder, m.Concatenator)))
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
	if m.IsPublished() {
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
	if m.IsPublished() {
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
	ac := c.MustGet("AerospikeClient").(*db.AerospikeClient)
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
	if m.IsPublished() {
		utils.ResponseError(c, http.StatusBadRequest, errors.New("you cannot upload data on already published models. stage it first"))
		return
	}

	// upload data from request itself
	if len(br.Data) > 0 && br.Data != nil {
		de, err := uploadDataDirectly(ac, br.Data, m)
		if err != nil {
			utils.ResponseError(c, http.StatusInternalServerError, err)
			return
		}
		utils.Response(c, http.StatusCreated, de)
		return
	}

	// truncate eventual old data
	if err := ac.TruncateSet(m.ComposeSetName()); err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	// upload data from S3 file
	bucket, key := StripS3URL(br.DataLocation)
	s := db.NewS3Client(&db.S3Bucket{Bucket: bucket, ACL: ""}, sess)

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
	go uploadDataFromFile(ac, f, m, batchID)

	utils.Response(c, http.StatusCreated, &BatchBulkResponse{BatchID: batchID})
}

// StripS3URL returns the bucket and the key from a s3 url location
func StripS3URL(URL string) (string, string) {
	bucketTmp := strings.Replace(URL, "s3://", "", -1)

	bucket := bucketTmp[:strings.IndexByte(bucketTmp, '/')]
	key := strings.TrimPrefix(URL, fmt.Sprintf("s3://%s/", bucket))

	return bucket, key
}

func uploadDataDirectly(ac *db.AerospikeClient, bd []BatchData, m *models.Model) (*BatchResponse, error) {
	var ln, ne int = 0, 0
	var vl bool = false
	var lineErrors []models.LineError

	setName := m.ComposeSetName()

	// check upfront if signal validation is required
	if requireSignalFormat(m) {
		vl = true
	}

	for _, data := range bd {
		for sig, recommendedItems := range data {
			ln++

			// validate if required
			if vl && !correctSignalFormat(m, sig) {
				ne++
				if ln <= maxErrorLines {
					msg := fmt.Sprintf("wrong format, the expected signal format must be %s", strings.Join(m.SignalOrder, m.Concatenator))
					lineErrors = append(lineErrors, models.LineError{strconv.Itoa(ln): msg})
				}
				continue
			}

			// upload to Aerospike
			if err := ac.AddOne(setName, sig, binKey, recommendedItems); err != nil {
				return nil, err
			}
		}
	}

	return &BatchResponse{
		NumberOfLines: strconv.Itoa(ln),
		ErrorRecords: DataUploadedError{
			Errors:              lineErrors,
			NumberOfLinesFailed: strconv.Itoa(ne),
		},
	}, nil
}

// BulkStatus defines the status of the batch upload from S3
type BulkStatus string

const (
	// BulkUploading represents the uploading status
	BulkUploading = "UPLOADING"
	// BulkSucceeded represents the succeeded status
	BulkSucceeded = "SUCCEEDED"
	// BulkPartialUpload represent the partial upload status
	BulkPartialUpload = "PARTIAL UPLOAD"
	// BulkFailed represents the failed status
	BulkFailed = "FAILED"
)

var (
	// MaxNumberOfWorkers is the max number of concurrent goroutines for uploading data
	MaxNumberOfWorkers = runtime.NumCPU()
)

func uploadDataFromFile(ac *db.AerospikeClient, file *io.ReadCloser, m *models.Model, batchID string) {
	start := time.Now()

	// write to Aerospike it's uploading
	if err := ac.AddOne(bulkStatusSetName, batchID, statusBinKey, BulkUploading); err != nil {
		log.Panic().Msg(err.Error())
	}

	setName := m.ComposeSetName()
	rd := bufio.NewReader(*file)
	rs := make(chan *models.RecordQueue)
	le := make(chan models.LineError)

	// create sync group
	wg := &sync.WaitGroup{}

	// fillup the channel with lines
	go func() {
		iterateFile(rd, setName, m, rs, le)
		close(rs)
		close(le)
	}()

	// store eventual errors
	go func() {
		if storeErrors(ac, batchID, le) > 0 {
			// write to Aerospike it partially uploded the data
			if err := ac.AddOne(bulkStatusSetName, batchID, statusBinKey, BulkPartialUpload); err != nil {
				log.Panic().Msg(err.Error())
			}
		} else {
			// write to Aerospike it succeeded
			if err := ac.AddOne(bulkStatusSetName, batchID, statusBinKey, BulkSucceeded); err != nil {
				log.Panic().Msg(err.Error())
			}
		}
	}()

	// consumes all the lines in parallel based on number of cpus
	wg.Add(MaxNumberOfWorkers)
	for i := 0; i < MaxNumberOfWorkers; i++ {
		go func() {
			uploadRecord(ac, batchID, rs)
			wg.Done()
		}()
	}

	// wait until done
	wg.Wait()

	elapsed := time.Since(start)
	log.Info().Msgf("Uploading took %s", elapsed)
}

func iterateFile(rd *bufio.Reader, setName string, m *models.Model, rs chan<- *models.RecordQueue, le chan<- models.LineError) {
	var ln int = 0
	var vl bool = false

	// check upfront if signal validation is required
	if requireSignalFormat(m) {
		vl = true
	}

	eof := false
	for !eof {
		line, err := rd.ReadString('\n')
		if err == io.EOF {
			eof = true
		}

		// string new-line character
		l := strings.TrimSuffix(line, "\n")

		// marshal the object
		var entry models.SingleEntry
		if err := json.Unmarshal([]byte(l), &entry); err != nil {
			le <- models.LineError{
				"lineRaw": line,
				"message": err.Error(),
			}
			continue
		}

		// validate signal format
		ln++
		if vl && !correctSignalFormat(m, entry.SignalID) {
			le <- models.LineError{
				"line":    strconv.Itoa(ln),
				"message": "signal not formatted correctly",
			}
			continue
		}

		// add to channel
		rs <- &models.RecordQueue{SetName: setName, Entry: entry, Error: nil}
	}
}

func uploadRecord(ac *db.AerospikeClient, batchID string, rs chan *models.RecordQueue) {
	// upload record to aerospike when it arrives
	for r := range rs {
		if err := ac.AddOne(r.SetName, r.Entry.SignalID, binKey, r.Entry.Recommended); err != nil {
			// write to Aerospike it failed
			if err := ac.AddOne(bulkStatusSetName, batchID, statusBinKey, BulkFailed); err != nil {
				// if this fails than since we cannot return the request to the user
				// we need to restart the application
				log.Panic().Msg(err.Error())
			}
			return
		}
	}
}

func storeErrors(ac *db.AerospikeClient, batchID string, le chan models.LineError) int {
	allErrors := []models.LineError{}
	i := 0
	for lineError := range le {
		if i < maxErrorLines {
			allErrors = append(allErrors, lineError)
			i++
			continue
		}
		break
	}
	// save to Aerospike the errors list
	if err := ac.AddOne(bulkStatusSetName, batchID, lineBinError, allErrors); err != nil {
		log.Panic().Msg(err.Error())
	}
	return i
}

// BatchStatusResponse is the response payload for getting the status of the bulk upload from S3
type BatchStatusResponse struct {
	Status string             `json:"status"`
	Errors []models.LineError `json:"errors"`
}

// BatchStatus returns the current status of the batch upload
func BatchStatus(c *gin.Context) {
	ac := c.MustGet("AerospikeClient").(*db.AerospikeClient)

	batchID := c.Param("id")

	r, err := ac.GetOne(bulkStatusSetName, batchID)
	if err != nil {
		utils.ResponseError(c, http.StatusNotFound, fmt.Errorf("batch job with ID %s not found", batchID))
		return
	}

	// save the status in a variable
	status := r.Bins[statusBinKey].(string)

	switch status {
	case BulkPartialUpload:
		errs := convertLineError(r.Bins[lineBinError])
		utils.Response(c, http.StatusOK, &BatchStatusResponse{Status: status, Errors: errs})
		break
	default:
		utils.Response(c, http.StatusOK, &BatchStatusResponse{Status: status})
	}
}

// The objects coming from Aerospike have type []interface{}. This function converts
// the Bins in the appropriate type for consistency
func convertLineError(bins interface{}) []models.LineError {
	var linesError []models.LineError
	newBins := bins.([]interface{})
	for _, bin := range newBins {
		b := bin.(map[interface{}]interface{})
		le := make(models.LineError)
		for k, v := range b {
			it := fmt.Sprintf("%v", k)
			score := fmt.Sprintf("%v", v)
			le[it] = score
		}
		linesError = append(linesError, le)
	}
	return linesError
}
