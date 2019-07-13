package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rtlnl/data-personalization-api/pkg/db"
)

// S3 is a middleware to inject the S3 client for accessing the bucket
func S3(bucket, region string) gin.HandlerFunc {
	client := db.NewS3Client(bucket, region)
	return func(c *gin.Context) {
		c.Set("S3Client", client)
		c.Next()
	}
}
