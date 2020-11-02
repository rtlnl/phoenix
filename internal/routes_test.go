package internal

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rtlnl/phoenix/middleware"
	"github.com/rtlnl/phoenix/pkg/db"
	"github.com/rtlnl/phoenix/utils"
)

var (
	testDBHost     = utils.GetEnv("DB_HOST", "127.0.0.1:6379")
	testDBPassword = utils.GetEnv("DB_PASSWORD", "qwerty")
	testBucket     = "test"
	testEndpoint   = "localhost:4572"
	testRegion     = "eu-west-1"
	testDisableSSL = true
)

var router *gin.Engine

func TestMain(m *testing.M) {
	tearUp()
	c := m.Run()
	tearDown()
	os.Exit(c)
}

func tearUp() {
	dbc, err := db.NewRedisClient(testDBHost, db.Password(testDBPassword))
	if err != nil {
		panic(err)
	}
	if err := dbc.Client.FlushAll().Err(); err != nil {
		panic(err)
	}

	// set server
	gin.SetMode(gin.TestMode)

	router = gin.Default()
	router.RedirectTrailingSlash = true

	router.Use(middleware.DB(dbc))
	router.Use(middleware.AWSSession(testRegion, testEndpoint, testDisableSSL))
	router.Use(middleware.NewWorker(dbc, "test-worker", "worker-queue"))

	// subscribe routes here due to multiple tests on the same endpoint
	// it avoids a panic error for registering the route multiple times
	router.POST("/v1/streaming", CreateStreaming)
	router.PUT("/v1/streaming", UpdateStreaming)
	router.DELETE("/v1/streaming", DeleteStreaming)
	router.DELETE("/v1/streaming/recommendation", DeleteRecommendation)

	router.POST("/v1/batch", Batch)
	router.GET("/v1/batch/status/:id", BatchStatus)

	// Management Routes
	mg := router.Group("/v1/management")

	// Container routes
	mc := mg.Group("/containers")
	mc.GET("/", GetContainer)
	mc.POST("/", CreateContainer)
	mc.DELETE("/", EmptyContainer)
	mc.PUT("/link-model", LinkModel)

	// Model routes
	mm := mg.Group("/models")
	mm.GET("/", GetModel)
	mm.POST("/", CreateModel)
	mm.DELETE("/", EmptyModel)
	mm.GET("/preview", GetDataPreview)
}

func tearDown() {
	dbc, err := db.NewRedisClient(testDBHost, db.Password(testDBPassword))
	if err != nil {
		panic(err)
	}
	if err := dbc.Client.FlushAll().Err(); err != nil {
		panic(err)
	}
}

func GetTestRedisClient() (db.DB, func()) {
	dbc, err := db.NewRedisClient(testDBHost, db.Password(testDBPassword))
	if err != nil {
		panic(err)
	}
	return dbc, func() {
		err := dbc.Close()
		if err != nil {
			panic(err)
		}
	}
}

// MockRequest will send a request to the server. Used for testing purposes
func MockRequest(method, path string, body io.Reader) (int, *bytes.Buffer, error) {
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return -1, nil, err
	}

	// Create a response recorder so you can inspect the response
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	return w.Code, w.Body, nil
}
