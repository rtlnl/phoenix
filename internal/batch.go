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
	"github.com/rtlnl/phoenix/utils"
	"github.com/rtlnl/phoenix/models"
	"github.com/rtlnl/phoenix/pkg/db"
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
	// FlushIntervalInSec is the amount of time before executing the Pipeline in case the buffer is not full
	FlushIntervalInSec = 30
	// MaxNumberOfCommandsInPipeline is the amount of commands that the Pipeline can executes in one step
	MaxNumberOfCommandsInPipeline = 10000
)

// BatchOperator is the object responsible for uploading data in batch to Database
type BatchOperator struct {
	DBClient db.DB
	Model    models.Model
}

// NewBatchOperator returns the object responsible for uploading the data in batch to Database
func NewBatchOperator(dbc db.DB, m models.Model) *BatchOperator {
	return &BatchOperator{
		DBClient: dbc,
		Model:    m,
	}
}

// UploadDataFromFile reads from a file and upload line-by-line to Database on a particular BatchID
func (bo *BatchOperator) UploadDataFromFile(file *io.ReadCloser, batchID string) {
	start := time.Now()

	// write to DB that it's uploading
	if err := bo.DBClient.AddOne(tableBulkStatus, batchID, BulkUploading); err != nil {
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
			// write to DB that it partially uploaded the data
			if err := bo.DBClient.AddOne(tableBulkStatus, batchID, BulkPartialUpload); err != nil {
				log.Panic().Msg(err.Error())
			}
		} else {
			// write to DB that it succeeded
			if err := bo.DBClient.AddOne(tableBulkStatus, batchID, BulkSucceeded); err != nil {
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
	log.Debug().Msgf("Uploading took %s", elapsed)
}

// UploadDataDirectly does an insert directly to Database
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

			// upload to DB
			ser, err := utils.SerializeObject(recommendedItems)
			if err != nil {
				log.Error().Msgf("could not serialize recommended object. error: %s", err.Error())
			}
			if err := bo.DBClient.AddOne(bo.Model.Name, sig, ser); err != nil {
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
		rs <- &models.RecordQueue{Table: setName, Entry: entry, Error: nil}
	}
}

// UploadRecord store each message from the channel to DB
func (bo *BatchOperator) UploadRecord(batchID string, rs chan *models.RecordQueue) {
	var buffer []string
	var fflush time.Time

	flushInterval := time.Duration(FlushIntervalInSec) * time.Second
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

Loop:
	for {
		if buffer == nil {
			buffer = make([]string, 0)
		}

		// either wait for message or grab the ticker
		select {
		case <-ticker.C:
			// Refresh pipe
			tt := time.Now()
			if tt.After(fflush) {
				log.Debug().Msg("Force flush (interval) triggered")
				// executes the pipeline
				bo.flushPipeline(batchID)
				// reset buffer
				buffer = nil
				fflush = tt.Add(flushInterval)
				log.Debug().Msg("Force flush (interval) finished")
			}
		case r := <-rs:
			// channel closed, finishing up
			if r == nil {
				log.Debug().Msg("Force flush (closed) triggered")
				// executes the pipeline
				bo.flushPipeline(batchID)
				// reset buffer
				buffer = nil
				log.Debug().Msg("Force flush (closed) finished")
				break Loop
			}

			// received message, continuing
			ser, err := utils.SerializeObject(r.Entry.Recommended)
			if err != nil {
				log.Error().Msgf("cold not serialize recommendations. error: %s", err.Error())
				continue
			}
			bo.DBClient.PipelineAddOne(r.Table, r.Entry.SignalID, ser)

			// append to buffer
			buffer = append(buffer, ser)

			if len(buffer) > MaxNumberOfCommandsInPipeline {
				log.Debug().Msg("Force flush (filled) triggered")
				// executes the pipeline
				bo.flushPipeline(batchID)
				// reset buffer
				buffer = nil
				log.Debug().Msg("Force flush (filled) finished")
			}
		}
	}
}

// flushPipeline executes the pipeline
func (bo *BatchOperator) flushPipeline(batchID string) {
	if err := bo.DBClient.PipelineExec(); err != nil {
		// write to DB that it failed
		if err := bo.DBClient.AddOne(tableBulkStatus, batchID, BulkFailed); err != nil {
			log.Error().Msg(err.Error())
		}
		log.Error().Msg(err.Error())
	}
}

// StoreErrors stores the errors in Database from the channel in input
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
	// save to DB the errors list
	ser, err := utils.SerializeObject(allErrors)
	if err != nil {
		log.Error().Msgf("could not serialize error objects. error: %s", err.Error())
	}
	if err := bo.DBClient.AddOne(tableBulkErrors, batchID, ser); err != nil {
		log.Panic().Msg(err.Error())
	}
	return i
}
