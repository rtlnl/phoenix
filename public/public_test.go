package public

import (
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rtlnl/phoenix/middleware"
	"github.com/rtlnl/phoenix/pkg/logs"
	"github.com/stretchr/testify/assert"
)

func TestNewPublicAPI(t *testing.T) {
	port, _ := strconv.Atoi(testDBPort)
	rl := logs.NewStdoutLog()

	var middlewares []gin.HandlerFunc
	middlewares = append(middlewares, middleware.Aerospike(testDBHost, testNamespace, port))
	middlewares = append(middlewares, middleware.RecommendationLogs(rl))

	p, err := NewPublicAPI(middlewares...)
	if err != nil {
		t.Fail()
	}

	assert.NotNil(t, p)
}
