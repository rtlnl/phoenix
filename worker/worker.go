package worker

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/adjust/rmq"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/rs/zerolog/log"
	"github.com/rtlnl/phoenix/models"
	"github.com/rtlnl/phoenix/pkg/aws"
	"github.com/rtlnl/phoenix/pkg/batch"
	"github.com/rtlnl/phoenix/pkg/db"
)

const (
	consumerName = "phoenix-consumer"
	consumerTag  = "phoenix-consumer-tag"
	unackedLimit = 10
	pollDuration = 500 * time.Millisecond
)

// Worker encapsulate the queueing system
type Worker struct {
	Queue      rmq.Queue
	Consumer   TaskConsumer
	AWSSession *session.Session
}

// TaskConsumer is the objecy around the Consumer interface
type TaskConsumer struct {
	name   string
	count  int
	before time.Time
}

// TaskPayload is the struct that contains the payload for consuming the task
type TaskPayload struct {
	DBURL        string `json:"db_url"`
	AWSRegion    string `json:"aws_region"`
	S3Endpoint   string `json:"s3_endpoint"`
	S3DisableSSL bool   `json:"s3_disable_ssl"`
	S3Bucket     string `json:"s3_bucket"`
	S3Key        string `json:"s3_key"`
	ModelName    string `json:"model_name"`
	BatchID      string `json:"batch_id"`
}

// New creates a new worker object
func New(broker, workerName, queueName string) (*Worker, error) {
	connection := rmq.OpenConnection(workerName, "tcp", broker, 1)
	queue := connection.OpenQueue(queueName)
	cs := TaskConsumer{
		name:   consumerName,
		count:  0,
		before: time.Now(),
	}
	return &Worker{Queue: queue, Consumer: cs}, nil
}

// Consume will execute the operation of batch uploading the data in Redis
func (c TaskConsumer) Consume(delivery rmq.Delivery) {
	var task *TaskPayload
	if err := json.Unmarshal([]byte(delivery.Payload()), &task); err != nil {
		delivery.Reject()
		return
	}

	sess := aws.NewAWSSession(task.AWSRegion, task.S3Endpoint, task.S3DisableSSL)
	s := db.NewS3Client(&db.S3Bucket{Bucket: task.S3Bucket, ACL: ""}, sess)

	// check if file exists
	if s.ExistsObject(task.S3Key) == false {
		log.Error().Msgf("key %s not founds in S3", task.S3Key)
		return
	}

	// download the file
	f, err := s.GetObject(task.S3Key)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}

	// create batch operator
	dbc, err := db.NewRedisClient(task.DBURL)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}

	// get the model
	m, err := models.GetModel(task.ModelName, dbc)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}

	bo := batch.NewOperator(dbc, m)
	if err := bo.UploadDataFromFile(f, task.BatchID); err != nil {
		delivery.Reject()
		return
	}
	// message processed correctly
	delivery.Ack()
}

// Consume instructs the worker to consuming the messages
func (w *Worker) Consume() error {
	if w.Queue.StartConsuming(unackedLimit, pollDuration) == false {
		return errors.New("could not start consuming messages")
	}

	res := w.Queue.AddConsumer(consumerTag, w.Consumer)
	log.Info().Msgf("worker added a new consumer %s", res)

	return nil
}
