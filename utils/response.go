package utils

import (
	"github.com/gin-gonic/gin"
)

// RespError represents the response structu when an error occurs
type RespError struct {
	Message string `json:"error" example:"status bad request"`
}

// ResponseError will return the error message
func ResponseError(c *gin.Context, status int, err error) {
	er := RespError{
		Message: err.Error(),
	}
	c.JSON(status, er)
}

// Response will return a response for the payload passed as argument
func Response(c *gin.Context, code int, payload interface{}) {
	c.JSON(code, payload)
}
