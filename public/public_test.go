package public

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rtlnl/phoenix/middleware"
	"github.com/rtlnl/phoenix/pkg/db"
	"github.com/rtlnl/phoenix/pkg/logs"
	"github.com/stretchr/testify/assert"
)

func TestNewPublicAPI(t *testing.T) {
	rl := logs.NewStdoutLog()

	// instantiate Redis client
	redisClient, err := db.NewRedisClient(testDBHost, nil)
	if err != nil {
		panic(err)
	}

	var middlewares []gin.HandlerFunc
	middlewares = append(middlewares, middleware.DB(redisClient))
	middlewares = append(middlewares, middleware.RecommendationLogs(rl))

	p, err := NewPublicAPI(middlewares...)
	if err != nil {
		t.Fail()
	}

	assert.NotNil(t, p)
}
