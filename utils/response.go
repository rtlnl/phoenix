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

// Resp represents the response struct when it's not an error
type Resp struct {
	Data interface{} `json:"data"`
}

// ResponseError will return the error message
func ResponseError(c *gin.Context, status int, err error) {
	er := RespError{
		Code:    status,
		Message: err.Error(),
	}
	c.JSON(status, er)
}

// Response will return a response wrapped around the 'resp' struct
func Response(c *gin.Context, code int, payload interface{}) {
	b, err := json.Marshal(&Resp{Data: payload})
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	r := string(b)
	c.String(code, r)
}
