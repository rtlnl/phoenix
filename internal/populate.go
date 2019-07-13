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

// Populate will take care of populating the personalized content for all the users
func Populate(c *gin.Context) {
	rc := c.MustGet("RedisClient").(*db.RedisClient)
	sc := c.MustGet("S3Client").(*db.S3Client)

	d := time.Now()
	today := d.Format("2006-01-02T15:04:05")

	u, err := uuid.NewUUID()
	if err != nil {
		utils.ResponseError(c, http.StatusInternalServerError, err)
		return
	}

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

/*

Key for history => history:YYYYMMDD -> UUID
Key for personalization => UUID:user:ID -> [ ... ]

	1. IF history not exists:
		1.1 create history with current date + uuid
		1.2 ELSE add current date + uuid

	2. Read file from S3
		2.1 Line by line add entry in Redis as: UUID_UserID: [it_1, it_2, ..., it_n]

	3. IF fails:
		3.1 it should not fail - problem with the data

*/
