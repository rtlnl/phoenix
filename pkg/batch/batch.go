package batch

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog/log"
	"github.com/rtlnl/phoenix/models"
	"github.com/rtlnl/phoenix/pkg/cache"
	"github.com/rtlnl/phoenix/pkg/db"
	"github.com/rtlnl/phoenix/utils"
)

// BulkStatus defines the status of the batch upload from S3
type BulkStatus string

const (
	// BulkQueued represents the status when the batch operation is queued
	BulkQueued = "QUEUED"
	// BulkUploading represents the uploading status
	BulkUploading = "UPLOADING"
	// BulkSucceeded represents the succeeded status
	BulkSucceeded = "SUCCEEDED"
	// BulkPartialUpload represent the partial upload status
	BulkPartialUpload = "PARTIAL UPLOAD"
	// BulkFailed represents the failed status
	BulkFailed = "FAILED"
	// TableBulkStatus is the name of the table for storing all the batchIDs
	TableBulkStatus = "bulkStatus"
	// TableBulkErrors is the name of the error tables for storing all the errors of a specific batch
	TableBulkErrors = "bulkErrors"
	// max number of Errors that will be stored in DB
	maxErrorLines = 50
)

var (
	// MaxNumberOfWorkers is the max number of concurrent goroutines for uploading data
	MaxNumberOfWorkers = runtime.NumCPU()
	// FlushIntervalInSec is the amount of time before executing the Pipeline in case the buffer is not full
	FlushIntervalInSec = 10
	// MaxNumberOfCommandsInPipeline is the amount of commands that the Pipeline can executes in one step
	MaxNumberOfCommandsInPipeline = 10000
)

// Data is the object representing the content of the data parameter in the batch request
type Data map[string][]models.ItemScore

// DataUploadedError is the response payload when the batch upload failed
type DataUploadedError struct {
	NumberOfLinesFailed string             `json:"numberoflinesfailed" description:"total count of lines that were not uploaded"`
	Errors              []models.LineError `json:"error" description:"errors found"`
}

// Operator is the object responsible for uploading data in batch to Database
type Operator struct {
	DBClient    db.DB
	CacheClient cache.Cache
	Model       models.Model
}

// NewOperator returns the object responsible for uploading the data in batch to Database
func NewOperator(dbc db.DB, m models.Model) *Operator {
	return &Operator{
		DBClient: dbc,
		Model:    m,
	}
}

// UploadDataFromFile reads from a file and upload line-by-line to Database on a particular BatchID
func (o *Operator) UploadDataFromFile(file *io.ReadCloser, batchID string) error {
	start := time.Now()

	// write to DB that it's uploading
	if err := o.SetStatus(batchID, BulkUploading); err != nil {
		return err
	}

	rd := bufio.NewReader(*file)
	rs := make(chan *models.RecordQueue)
	le := make(chan models.LineError)

	// create sync group
	wg := &sync.WaitGroup{}

	// fillup the channel with lines
	go func() {
		o.IterateFile(rd, o.Model.Name, rs, le)
		close(rs)
		close(le)
	}()

	// store eventual errors
	go func() {
		if o.StoreErrors(batchID, le) > 0 {
			// write to DB that it partially uploaded the data
			o.SetStatus(batchID, BulkPartialUpload)
		} else {
			// write to DB that it succeeded
			o.SetStatus(batchID, BulkSucceeded)
		}
	}()

	// consumes all the lines in parallel based on number of cpus
	wg.Add(MaxNumberOfWorkers)
	for i := 0; i < MaxNumberOfWorkers; i++ {
		go func() {
			o.UploadRecord(batchID, rs)
			wg.Done()
		}()
	}

	// wait until done
	wg.Wait()

	elapsed := time.Since(start)
	log.Info().Str("BATCH", fmt.Sprintf("upload in %s", elapsed))

	return nil
}

// UploadDataDirectly does an insert directly to Database
func (o *Operator) UploadDataDirectly(bd []Data) (string, DataUploadedError, error) {
	var ln, ne int = 0, 0
	var vl bool = false
	var lineErrors []models.LineError

	// check upfront if signal validation is required
	if o.Model.RequireSignalFormat() {
		vl = true
	}

	for _, data := range bd {
		for sig, recommendedItems := range data {
			ln++

			// validate if required
			if vl && !o.Model.CorrectSignalFormat(sig) {
				ne++
				if ln <= maxErrorLines {
					msg := fmt.Sprintf("wrong format, the expected signal format must be %s", strings.Join(o.Model.SignalOrder, o.Model.Concatenator))
					lineErrors = append(lineErrors, models.LineError{strconv.Itoa(ln): msg})
				}
				continue
			}

			// upload to DB
			ser, err := utils.SerializeObject(recommendedItems)
			if err != nil {
				log.Error().Msgf("could not serialize recommended object. error: %s", err.Error())
			}
			if err := o.DBClient.AddOne(o.Model.Name, sig, ser); err != nil {
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
func (o *Operator) IterateFile(rd *bufio.Reader, setName string, rs chan<- *models.RecordQueue, le chan<- models.LineError) {
	var ln int = 0
	vl := o.Model.RequireSignalFormat()

	for {
		line, err := rd.ReadString('\n')
		if err == io.EOF {
			break
		}

		// string new-line character
		l := strings.TrimSuffix(line, "\n")

		// marshal the object
		var entry models.SingleEntry
		if err := json.Unmarshal([]byte(l), &entry); err != nil {
			le <- models.LineError{
				"line":    strconv.Itoa(ln),
				"lineRaw": l,
				"message": spew.Sdump(entry),
			}
			log.Warn().Str("READ", fmt.Sprintf("could not serialize recommended object. error: %s", err.Error())).Str("LINE", line)
			continue
		}
		// validate signal format
		ln++
		if vl && !o.Model.CorrectSignalFormat(entry.SignalID) {
			le <- models.LineError{
				"line":    strconv.Itoa(ln),
				"message": "signal not formatted correctly",
			}
			log.Warn().Str("READ", "signal not formatted correctly").Str("SIGNAL", entry.SignalID).Str("LINE", line)
			continue
		}
		// add to channel
		rs <- &models.RecordQueue{Table: setName, Entry: entry, Error: nil}
	}
}

// UploadRecord store each message from the channel to DB
func (o *Operator) UploadRecord(batchID string, rs chan *models.RecordQueue) {
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
				o.flushPipeline(batchID)
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
				o.flushPipeline(batchID)
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
			o.DBClient.PipelineAddOne(r.Table, r.Entry.SignalID, ser)
			log.Info().Str("INSERT", fmt.Sprintf("signalId %s", r.Entry.SignalID)).Str("MODEL", o.Model.Name)

			// append to buffer
			buffer = append(buffer, ser)

			if len(buffer) > MaxNumberOfCommandsInPipeline {
				log.Debug().Msg("Force flush (filled) triggered")
				// executes the pipeline
				o.flushPipeline(batchID)
				// reset buffer
				buffer = nil
				log.Debug().Msg("Force flush (filled) finished")
			}
		}
	}
}

// flushPipeline executes the pipeline
func (o *Operator) flushPipeline(batchID string) {
	if err := o.DBClient.PipelineExec(); err != nil {
		log.Error().Msg(err.Error())
		// write to DB that it failed
		if err := o.DBClient.AddOne(TableBulkStatus, batchID, BulkFailed); err != nil {
			log.Error().Msg(err.Error())
		}
		log.Error().Msg(err.Error())
	}
}

// StoreErrors stores the errors in Database from the channel in input
func (o *Operator) StoreErrors(batchID string, le chan models.LineError) int {
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
	// save to DB the errors list if any
	if len(allErrors) > 0 {
		ser, err := utils.SerializeObject(allErrors)
		if err != nil {
			log.Error().Msgf("could not serialize error objects. error: %s", err.Error())
		}
		if err := o.DBClient.AddOne(TableBulkErrors, batchID, ser); err != nil {
			log.Panic().Msg(err.Error())
		}
	}
	return i
}

// SetStatus sets the status in the DB. The error message is logged only
func (o *Operator) SetStatus(batchID, status string) error {
	err := o.DBClient.AddOne(TableBulkStatus, batchID, status)
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	return err
}
