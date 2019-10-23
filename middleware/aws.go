package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rtlnl/phoenix/pkg/aws"
)

// AWSSession is a middleware to inject the AWSSession client for accessing the aws services
func AWSSession(region, endpoint string, disableSSL bool) gin.HandlerFunc {
	s := aws.NewAWSSession(region, endpoint, disableSSL)
	return func(c *gin.Context) {
		c.Set("AWSSession", s)
		c.Next()
	}
}
