package internal

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rtlnl/phoenix/middleware"
	"github.com/stretchr/testify/assert"
)

func TestNewInternalAPI(t *testing.T) {
	var middlewares []gin.HandlerFunc
	middlewares = append(middlewares, middleware.AWSSession(testBucket, testEndpoint, testDisableSSL))

	i, err := NewInternalAPI(middlewares...)
	if err != nil {
		t.Fail()
	}

	assert.NotNil(t, i)
}
