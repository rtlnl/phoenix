package internal

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/rtlnl/data-personalization-api/pkg/db"
	"github.com/rtlnl/data-personalization-api/utils"

	"github.com/gin-gonic/gin"
)

// AddPersonalizations will take care of populating the personalized content for all the users
func AddPersonalizations(c *gin.Context) {
	rc := c.MustGet("RedisClient").(*db.RedisClient)
	sc := c.MustGet("S3Client").(*db.S3Client)

	d := time.Now()
	today := d.Format("2006-01-02T15:04:05")

	u, err := uuid.NewUUID()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	// Key for history => history:YYYYMMDD -> UUID
	kh := fmt.Sprintf("history:%s", today)
	err = rc.SetValue(kh, u.String())
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	// get file from S3
	s3Key := fmt.Sprintf("content/%s/personalization.csv", today)
	f, err := sc.GetObject(s3Key)
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	// bulk import in Redis
	if rc.BulkImport(u, *f) != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	utils.Response(c, http.StatusCreated, "import succeeded")
}

// DeletePersonalizations will delete the personalized content of the previous day
func DeletePersonalizations(c *gin.Context) {

}
