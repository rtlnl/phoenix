package internal

import (
	"net/http"

	"github.com/rtlnl/data-personalization-api/pkg/db"
	"github.com/rtlnl/data-personalization-api/utils"

	"github.com/gin-gonic/gin"
)

// AddPersonalizations will take care of populating the personalized content for all the users
func AddPersonalizations(c *gin.Context) {
	_ = c.MustGet("AerospikeClient").(*db.AerospikeClient)
	_ = c.MustGet("S3Client").(*db.S3Client)

	utils.Response(c, http.StatusCreated, "import succeeded")
}

// DeletePersonalizations will delete the personalized content of the previous day
func DeletePersonalizations(c *gin.Context) {

}
