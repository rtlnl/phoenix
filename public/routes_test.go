package public

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/rtlnl/phoenix/middleware"
	"github.com/rtlnl/phoenix/models"
	"github.com/rtlnl/phoenix/pkg/db"
	"github.com/rtlnl/phoenix/utils"
)

var (
	testDBHost    = utils.GetEnv("DB_HOST", "127.0.0.1")
	testDBPort    = utils.GetEnv("DB_PORT", "3000")
	testNamespace = "test"
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

	router = gin.New()
	router.RedirectTrailingSlash = true

	p, _ := strconv.Atoi(testDBPort)
	router.Use(middleware.Aerospike(testDBHost, testNamespace, p))

	// subscribe route Recommend here due to multiple tests on this route
	// it avoids a panic error for registering the route multiple times
	router.GET("/recommend", Recommend)
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

	m, err := models.GetExistingModel(modelName, ac)
	if err != nil {
		t.Fatal(err)
	}

	sc := bufio.NewScanner(f)
	var entry models.SingleEntry

	i := 0
	for sc.Scan() {
		line := sc.Text() // GET the line string

		// marshal the object
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatal("fixtures contains the wrong type of json")
		}

		if err := ac.AddOne(m.Name, entry.SignalID, binKey, entry.Recommended); err != nil {
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
func CreateTestModel(t *testing.T, ac *db.AerospikeClient, modelName, concatenator string, signalType []string, publish bool) func() {
	m, _ := models.NewModel(modelName, concatenator, signalType, ac)
	if m == nil {
		m, _ = models.GetExistingModel(modelName, ac)
	}

	if publish {
		if err := m.PublishModel(ac); err != nil {
			t.FailNow()
		}
	}

	return func() {
		ac.TruncateSet(modelName)
	}
}

// CreateTestContainer returns a container and defer a truncate
func CreateTestContainer(t *testing.T, ac *db.AerospikeClient, publicationPoint, campaign string, modelsName []string) func() {
	c, _ := models.NewContainer(publicationPoint, campaign, modelsName, ac)
	if c == nil {
		c, _ = models.GetExistingContainer(publicationPoint, campaign, ac)
	}

	return func() {
		ac.TruncateSet(publicationPoint)
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

// MockRequestBenchmark will send a request to the server. Used for benchamrking purposes
func MockRequestBenchmark(b *testing.B, method, path string, body io.Reader) {
	req, _ := http.NewRequest(method, path, body)

	// Create a response recorder so you can inspect the response
	w := httptest.NewRecorder()

	// Perform the request
	b.StartTimer()
	router.ServeHTTP(w, req)
	b.StopTimer()
}
