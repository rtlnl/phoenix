package internal

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rtlnl/data-personalization-api/middleware"
	"github.com/rtlnl/data-personalization-api/models"
	"github.com/rtlnl/data-personalization-api/pkg/db"
	"github.com/rtlnl/data-personalization-api/utils"
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
	router.POST("/streaming", CreateStreaming)
	router.PUT("/streaming", UpdateStreaming)
	router.DELETE("/streaming", DeleteStreaming)

	router.POST("/batch", Batch)

	// Management Routes
	mm := router.Group("/management/model")
	mm.GET("", GetModel)
	mm.POST("", CreateModel)
	mm.DELETE("", EmptyModel)
	mm.POST("/publish", PublishModel)
}

func tearDown() {
	router = nil
}

func UploadTestData(t *testing.T, ac *db.AerospikeClient, testDataPath, modelName string) func() {
	f, err := os.OpenFile(testDataPath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)

	var entry models.SingleEntry

	i := 0
	for sc.Scan() {
		line := sc.Text() // GET the line string

		// marshal the object
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatal("fixtures contains the wrong type of json")
		}

		// need to break down the name of the model to fetch the actual object
		// from aerospike
		mn := strings.Split(modelName, "#")

		// TODO: this is slow! Need to find a smarter way. it works fine with small fixtures files
		m, err := models.GetExistingModel(mn[0], mn[1], ac)
		if err != nil {
			t.Fatal(err)
		}

		sn := m.ComposeSetName()
		if err := ac.AddOne(sn, entry.SignalID, binKey, entry.Recommended); err != nil {
			t.Fatal(err)
		}

		i++
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return func() { ac.TruncateSet(modelName) }
}

// GetTestAerospikeClient returns the client used for tests
func GetTestAerospikeClient() (*db.AerospikeClient, func()) {
	p, _ := strconv.Atoi(testDBPort)
	ac := db.NewAerospikeClient(testDBHost, testNamespace, p)

	return ac, func() { ac.Close() }
}

// CreateTestModel returns a model and defer a truncate
func CreateTestModel(t *testing.T, ac *db.AerospikeClient, publicationPoint, campaign, signalType string, publish bool) func() {
	m, _ := models.NewModel(publicationPoint, campaign, signalType, ac)
	if m == nil {
		m, _ = models.GetExistingModel(publicationPoint, campaign, ac)
	}

	if publish {
		if err := m.PublishModel(ac); err != nil {
			t.FailNow()
		}
	}

	return func() { ac.TruncateSet(publicationPoint) }
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
