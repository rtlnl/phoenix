package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rtlnl/phoenix/pkg/db"
	"github.com/rtlnl/phoenix/worker"
)

// NewWorker creates a new worker middleware
func NewWorker(rc *db.Redis, workerName, queueName string) gin.HandlerFunc {
	w, err := worker.New(rc.Client, workerName, queueName)
	if err != nil {
		panic(err)
	}
	return func(c *gin.Context) {
		c.Set("Worker", w)
		c.Next()
	}
}
