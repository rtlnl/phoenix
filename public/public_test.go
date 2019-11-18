package public

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rtlnl/phoenix/middleware"
	"github.com/rtlnl/phoenix/pkg/logs"
	"github.com/stretchr/testify/assert"
)

func TestNewPublicAPI(t *testing.T) {
	rl := logs.NewStdoutLog()

	// instantiate Redis client
	dbc, c := GetTestRedisClient()
	defer c()

	var middlewares []gin.HandlerFunc
	middlewares = append(middlewares, middleware.DB(dbc))
	middlewares = append(middlewares, middleware.RecommendationLogs(rl))

	p, err := NewPublicAPI(middlewares...)
	if err != nil {
		t.Fail()
	}

	assert.NotNil(t, p)
}
