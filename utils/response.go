package utils

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RespError represents the response structu when an error occurs
type RespError struct {
	Message string `json:"message" example:"status bad request"`
}

// ResponseError will return the error message
func ResponseError(c *gin.Context, status int, err error) {
	er := RespError{
		Message: err.Error(),
	}
	c.JSON(status, er) 
}

// Response will return a response for the payload passed as argument
// interface is like the object class in Java
func Response(c *gin.Context, code int, payload interface{}) {

	// b: array of bytes [slice] (check go slice) 
	b, err := json.Marshal(payload)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err)
		return
	}

	// Next block could be done with c.JSON
 
	// JSON string
	r := string(b)

	// set the header to format the response to json
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.String(code, r) // Wrapper function takes a string and compose the response 
}
