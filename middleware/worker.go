package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rtlnl/phoenix/worker"
)

// NewWorker creates a new worker middleware
func NewWorker(broker, workerName, queueName string) gin.HandlerFunc {
	w, err := worker.New(broker, workerName, queueName)
	if err != nil {
		panic(err)
	}
	return func(c *gin.Context) {
		c.Set("Worker", w)
		c.Next()
	}
}
