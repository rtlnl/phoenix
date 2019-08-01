package utils

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RespError represents the response structu when an error occurs
type RespError struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"status bad request"`
}

// ResponseError will return the error message
func ResponseError(c *gin.Context, status int, err error) {
	er := RespError{
		Code:    status,
		Message: err.Error(),
	}
	c.JSON(status, er)
}

// Response will return a response for the payload passed as argument
func Response(c *gin.Context, code int, payload interface{}) {
	b, err := json.Marshal(payload)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	r := string(b)
	c.String(code, r)
}
