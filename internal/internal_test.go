package internal

import (
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rtlnl/phoenix/middleware"
	"github.com/stretchr/testify/assert"
)

func TestNewInternalAPI(t *testing.T) {
	p, _ := strconv.Atoi(testDBPort)

	var middlewares []gin.HandlerFunc
	middlewares = append(middlewares, middleware.Aerospike(testDBHost, testNamespace, p))
	middlewares = append(middlewares, middleware.AWSSession(testBucket, testEndpoint, testDisableSSL))

	i, err := NewInternalAPI(middlewares...)
	if err != nil {
		t.Fail()
	}

	assert.NotNil(t, i)
}
