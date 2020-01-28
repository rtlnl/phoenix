package worker

import (
	"github.com/rs/zerolog/log"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// ConsumerTag represents the unique value of the worker
	ConsumerTag = "phoenix_consumer"
)

var (
	taskStartedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "task_started",
		Help: "Count of all Started Tasks",
	}, []string{"code", "method"})
	taskSucceededTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "task_completed",
		Help: "Count of all Succeeded Tasks",
	}, []string{"code", "method"})
	taskFailedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "task_failed",
		Help: "Count of all Failed Tasks",
	}, []string{"code", "method"})
)

// Worker encapsulate the queueing system
type Worker struct {
	Server *machinery.Server
}

// NewWorker creates a new worker object
func NewWorker(broker, resultBackend string) (*Worker, error) {
	cnf := config.Config{
		Broker:        broker,
		ResultBackend: resultBackend,
	}

	// Register the tasks with Prometheus's default registry.
	prometheus.MustRegister(taskFailedTotal)
	prometheus.MustRegister(taskStartedTotal)
	prometheus.MustRegister(taskSucceededTotal)
	// Add Go module build info.
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())

	// Create server instance
	server, err := machinery.NewServer(&cnf)
	if err != nil {
		return nil, err
	}

	return &Worker{Server: server}, nil
}

// RegisterTasks register the tasks to the worker
func (w *Worker) RegisterTasks(tasks map[string]interface{}) error {
	return w.Server.RegisterTasks(tasks)
}

// Launch instructs the worker to start
func (w *Worker) Launch() error {

	// The second argument is a consumer tag
	// Ideally, each worker should have a unique tag (worker1, worker2 etc)
	worker := w.Server.NewWorker(ConsumerTag, 0)

	// Here we inject some custom code for error handling,
	// start and end of task hooks, useful for metrics for example.
	errorhandler := func(err error) {
		taskFailedTotal.WithLabelValues("FAILED").Inc()
		log.Error().Msg(err.Error())
	}

	pretaskhandler := func(signature *tasks.Signature) {
		taskStartedTotal.WithLabelValues("STARTED").Inc()
		log.Debug().Msgf("task %s started...", signature.Name)
	}

	posttaskhandler := func(signature *tasks.Signature) {
		taskSucceededTotal.WithLabelValues("SUCCEEDED").Inc()
		log.Debug().Msgf("task %s completed successfully", signature.Name)
		taskStartedTotal.WithLabelValues("STARTED").Add(-1)
	}

	worker.SetPostTaskHandler(posttaskhandler)
	worker.SetErrorHandler(errorhandler)
	worker.SetPreTaskHandler(pretaskhandler)

	return worker.Launch()
}
