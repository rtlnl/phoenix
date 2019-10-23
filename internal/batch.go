package internal

import (
	"bufio"
	"fmt"
	"io"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/rtlnl/data-personalization-api/models"
	"github.com/rtlnl/data-personalization-api/pkg/db"
)

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

// BatchOperator is the object responsible for uploading data in batch to Aerospike
type BatchOperator struct {
	AeroClient *db.AerospikeClient
	Model      *models.Model
}

// NewBatchOperator returns the object responsible for uploading the data in batch to Aerospike
func NewBatchOperator(ac *db.AerospikeClient, m *models.Model) *BatchOperator {
	return &BatchOperator{
		AeroClient: ac,
		Model:      m,
	}
}

// UploadDataFromFile reads from a file and upload line-by-line to aerospike on a particular BatchID
func (bo *BatchOperator) UploadDataFromFile(file *io.ReadCloser, batchID string) {
	start := time.Now()

	// write to Aerospike it's uploading
	if err := bo.AeroClient.AddOne(bulkStatusSetName, batchID, statusBinKey, BulkUploading); err != nil {
		log.Panic().Msg(err.Error())
	}

	rd := bufio.NewReader(*file)
	rs := make(chan *models.RecordQueue)
	le := make(chan models.LineError)

	// create sync group
	wg := &sync.WaitGroup{}

	// fillup the channel with lines
	go func() {
		bo.IterateFile(rd, bo.Model.Name, rs, le)
		close(rs)
		close(le)
	}()

	// store eventual errors
	go func() {
		if bo.StoreErrors(batchID, le) > 0 {
			// write to Aerospike it partially uploded the data
			if err := bo.AeroClient.AddOne(bulkStatusSetName, batchID, statusBinKey, BulkPartialUpload); err != nil {
				log.Panic().Msg(err.Error())
			}
		} else {
			// write to Aerospike it succeeded
			if err := bo.AeroClient.AddOne(bulkStatusSetName, batchID, statusBinKey, BulkSucceeded); err != nil {
				log.Panic().Msg(err.Error())
			}
		}
	}()

	// consumes all the lines in parallel based on number of cpus
	wg.Add(MaxNumberOfWorkers)
	for i := 0; i < MaxNumberOfWorkers; i++ {
		go func() {
			bo.UploadRecord(batchID, rs)
			wg.Done()
		}()
	}

	// wait until done
	wg.Wait()

	elapsed := time.Since(start)
	log.Info().Msgf("Uploading took %s", elapsed)
}

// UploadDataDirectly does an insert directly to aerospike
func (bo *BatchOperator) UploadDataDirectly(bd []BatchData) (string, DataUploadedError, error) {
	var ln, ne int = 0, 0
	var vl bool = false
	var lineErrors []models.LineError

	// check upfront if signal validation is required
	if bo.Model.RequireSignalFormat() {
		vl = true
	}

	for _, data := range bd {
		for sig, recommendedItems := range data {
			ln++

			// validate if required
			if vl && !bo.Model.CorrectSignalFormat(sig) {
				ne++
				if ln <= maxErrorLines {
					msg := fmt.Sprintf("wrong format, the expected signal format must be %s", strings.Join(bo.Model.SignalOrder, bo.Model.Concatenator))
					lineErrors = append(lineErrors, models.LineError{strconv.Itoa(ln): msg})
				}
				continue
			}

			// upload to Aerospike
			if err := bo.AeroClient.AddOne(bo.Model.Name, sig, binKey, recommendedItems); err != nil {
				return "", DataUploadedError{}, err
			}
		}
	}

	// compose object for errors
	due := DataUploadedError{
		Errors:              lineErrors,
		NumberOfLinesFailed: strconv.Itoa(ne),
	}

	return strconv.Itoa(ln), due, nil
}

// IterateFile will iterate each line in the reader object and push messages in the channels
func (bo *BatchOperator) IterateFile(rd *bufio.Reader, setName string, rs chan<- *models.RecordQueue, le chan<- models.LineError) {
	var ln int = 0
	var vl bool = false

	// check upfront if signal validation is required
	if bo.Model.RequireSignalFormat() {
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
		if vl && !bo.Model.CorrectSignalFormat(entry.SignalID) {
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

// UploadRecord store each message from the channel to Aerospike
func (bo *BatchOperator) UploadRecord(batchID string, rs chan *models.RecordQueue) {
	// upload record to aerospike when it arrives
	for r := range rs {
		if err := bo.AeroClient.AddOne(r.SetName, r.Entry.SignalID, binKey, r.Entry.Recommended); err != nil {
			// write to Aerospike it failed
			if err := bo.AeroClient.AddOne(bulkStatusSetName, batchID, statusBinKey, BulkFailed); err != nil {
				// if this fails than since we cannot return the request to the user
				// we need to restart the application
				log.Panic().Msg(err.Error())
			}
			return
		}
	}
}

// StoreErrors stores the errors in Aerospike from the channel in input
func (bo *BatchOperator) StoreErrors(batchID string, le chan models.LineError) int {
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
	if err := bo.AeroClient.AddOne(bulkStatusSetName, batchID, lineBinError, allErrors); err != nil {
		log.Panic().Msg(err.Error())
	}
	return i
}
