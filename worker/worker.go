package worker

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/adjust/rmq/v3"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-redis/redis/v7"
	"github.com/rs/zerolog/log"
	"github.com/rtlnl/phoenix/models"
	"github.com/rtlnl/phoenix/pkg/aws"
	"github.com/rtlnl/phoenix/pkg/batch"
	"github.com/rtlnl/phoenix/pkg/db"
)

const (
	// WorkerLockKey is the key for locking/unlocking
	WorkerLockKey = "worker:lock"
	consumerName  = "phoenix-consumer"
	consumerTag   = "phoenix-consumer-tag"
	unackedLimit  = 10
	pollDuration  = 15 * time.Second
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
	DBPassword   string `json:"db_password"`
	AWSRegion    string `json:"aws_region"`
	S3Endpoint   string `json:"s3_endpoint"`
	S3DisableSSL bool   `json:"s3_disable_ssl"`
	S3Bucket     string `json:"s3_bucket"`
	S3Key        string `json:"s3_key"`
	ModelName    string `json:"model_name"`
	BatchID      string `json:"batch_id"`
}

// New creates a new worker object
func New(rc *redis.Client, workerName, queueName string) (*Worker, error) {
	connection, err := rmq.OpenConnectionWithRedisClient(workerName, rc, nil)
	if err != nil {
		return nil, err
	}

	queue, err := connection.OpenQueue(queueName)
	if err != nil {
		return nil, err
	}

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

	// create batch operator
	dbc, err := db.NewRedisClient(task.DBURL, db.Password(task.DBPassword))
	if err != nil {
		log.Error().Msg(err.Error())
		delivery.Reject()
		return
	}

	// get the model
	m, err := models.GetModel(task.ModelName, dbc)
	if err != nil {
		log.Error().Msg(err.Error())
		delivery.Reject()
		return
	}

	// create batch operator
	bo := batch.NewOperator(dbc, m)

	// check if file exists
	if s.ExistsObject(task.S3Key) == false {
		bo.SetStatus(task.BatchID, batch.BulkFailed)
		log.Error().Msgf("key %s not founds in S3", task.S3Key)
		delivery.Reject()
		return
	}

	// download the file
	f, err := s.GetObject(task.S3Key)
	if err != nil {
		bo.SetStatus(task.BatchID, batch.BulkFailed)
		log.Error().Msg(err.Error())
		delivery.Reject()
		return
	}

	if err := bo.UploadDataFromFile(f, task.BatchID); err != nil {
		bo.SetStatus(task.BatchID, batch.BulkFailed)
		delivery.Reject()
		return
	}
	// message processed correctly
	delivery.Ack()
}

// Consume instructs the worker to consuming the messages
func (w *Worker) Consume() error {
	if err := w.Queue.StartConsuming(unackedLimit, pollDuration); err != nil {
		return errors.New("could not start consuming messages")
	}
	w.Queue.AddConsumer(consumerTag, w.Consumer)
	return nil
}

// Publish publishes the message in the queue
func (w *Worker) Publish(tp *TaskPayload) error {
	b, err := json.Marshal(tp)
	if err != nil {
		return err
	}

	if err := w.Queue.PublishBytes(b); err != nil {
		return errors.New("could not publish message to queue")
	}
	return nil
}

// Close closes the queue
func (w *Worker) Close() {
	w.Queue.StopConsuming()
}
