package public

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rtlnl/phoenix/middleware"
	"github.com/rtlnl/phoenix/models"
	"github.com/rtlnl/phoenix/pkg/db"
	"github.com/rtlnl/phoenix/pkg/logs"
	"github.com/rtlnl/phoenix/utils"
)

var (
	testDBHost    = utils.GetEnv("DB_HOST", "127.0.0.1:6379")
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

	// instantiate Redis client
	redisClient, err := db.NewRedisClient(testDBHost)
	if err != nil {
		panic(err)
	}

	router.Use(middleware.DB(redisClient))
	router.Use(middleware.RecommendationLogs(logs.NewStdoutLog()))

	// subscribe route Recommend here due to multiple tests on this route
	// it avoids a panic error for registering the route multiple times
	router.GET("/v1/recommend", Recommend)
}

func tearDown() {
	router = nil

	dbc, c := GetTestRedisClient()
	defer c()

	if err := dbc.DropTable("tableModels"); err != nil {
		panic(err.Error())
	}

	if err := dbc.DropTable("tableContainers"); err != nil {
		panic(err.Error())
	}
}

func GetTestRedisClient() (db.DB, func()) {
	dbc, err := db.NewRedisClient(testDBHost)
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

func UploadTestData(t *testing.T, dbc db.DB, testDataPath, modelName string) {
	f, err := os.OpenFile(testDataPath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	m, err := models.GetModel(modelName, dbc)
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

		ser, err := utils.SerializeObject(entry.Recommended)
		if err != nil {
			t.FailNow()
		}

		if err := dbc.AddOne(m.Name, entry.SignalID, ser); err != nil {
			t.Fatal(err)
		}
		i++
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
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
