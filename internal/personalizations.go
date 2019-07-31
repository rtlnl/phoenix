package internal

import (
	"net/http"

	"github.com/rtlnl/data-personalization-api/pkg/db"
	"github.com/rtlnl/data-personalization-api/utils"

	"github.com/gin-gonic/gin"
)

// Streaming will upload in a streaming fashion a set of data
func Streaming(c *gin.Context) {
	_ = c.MustGet("AerospikeClient").(*db.AerospikeClient)
	_ = c.MustGet("S3Client").(*db.S3Client)

	utils.Response(c, http.StatusCreated, "import succeeded")
}

// Batch will upload in batch a set to the database
func Batch(c *gin.Context) {

}
