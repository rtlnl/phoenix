package internal

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rtlnl/phoenix/middleware"
	"github.com/rtlnl/phoenix/models"
	"github.com/rtlnl/phoenix/pkg/db"
	"github.com/rtlnl/phoenix/utils"
)

var (
	testDBHost     = utils.GetEnv("DB_HOST", "127.0.0.1")
	testDBPort     = utils.GetEnv("DB_PORT", "3000")
	testNamespace  = "test"
	testBucket     = "test"
	testRegion     = "eu-west-1"
	testEndpoint   = "localhost:4572"
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

	gin.SetMode(gin.TestMode)

	router = gin.Default()
	router.RedirectTrailingSlash = true

	p, _ := strconv.Atoi(testDBPort)
	router.Use(middleware.Aerospike(testDBHost, testNamespace, p))
	router.Use(middleware.AWSSession(testRegion, testEndpoint, testDisableSSL))

	// subscribe routes here due to multiple tests on the same endpoint
	// it avoids a panic error for registering the route multiple times
	router.POST("/v1/streaming", CreateStreaming)
	router.PUT("/v1/streaming", UpdateStreaming)
	router.DELETE("/v1/streaming", DeleteStreaming)

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
	mm.POST("/publish", PublishModel)
	mm.POST("/stage", StageModel)
}

func tearDown() {
	router = nil

	ac, _ := GetTestAerospikeClient()
	ac.TruncateNamespace("test")
	time.Sleep(2 * time.Second)
}

// GetTestAerospikeClient returns the client used for tests
func GetTestAerospikeClient() (*db.AerospikeClient, func()) {
	p, _ := strconv.Atoi(testDBPort)
	ac := db.NewAerospikeClient(testDBHost, testNamespace, p)

	return ac, func() { ac.Close() }
}

// CreateTestModel returns a model and defer a truncate
func CreateTestModel(t *testing.T, ac *db.AerospikeClient, name, concatenator string, signalType []string, publish bool) func() {
	m, _ := models.NewModel(name, concatenator, signalType, ac)
	if m == nil {
		m, _ = models.GetExistingModel(name, ac)
	}

	if publish {
		if err := m.PublishModel(ac); err != nil {
			t.FailNow()
		}
	}

	return func() {
		ac.TruncateSet(name)
		time.Sleep(2 * time.Second)
	}
}

// CreateTestContainer returns a container and defer a truncate
func CreateTestContainer(t *testing.T, ac *db.AerospikeClient, publicationPoint, campaign string, modelsName []string) func() {
	c, err := models.NewContainer(publicationPoint, campaign, modelsName, ac)
	if err != nil {
		t.Errorf("CreateTestContainer has an error %s", err.Error())
	}
	if c == nil {
		c, _ = models.GetExistingContainer(publicationPoint, campaign, ac)
	}

	return func() {
		ac.TruncateSet(publicationPoint)
		time.Sleep(2 * time.Second)
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
