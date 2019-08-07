package public

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/rtlnl/data-personalization-api/models"

	"github.com/gin-gonic/gin"
	"github.com/rtlnl/data-personalization-api/middleware"
	"github.com/rtlnl/data-personalization-api/pkg/db"
)

const (
	testDBHost    = "127.0.0.1"
	testDBPort    = 3000
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

	router = gin.Default()
	router.RedirectTrailingSlash = true

	// Load fixtures
	loadFixtures()

	router.Use(middleware.Aerospike(testDBHost, testNamespace, testDBPort))

	// subscribe route Recommend here due to multiple tests on this route
	// it avoids a panic error for registering the route multiple times
	router.POST("/recommend", Recommend)
}

func tearDown() {
	router = nil
}

func loadFixtures() {
	ac := db.NewAerospikeClient(testDBHost, testNamespace, testDBPort)

	// load fixtures here
	// model
	if err := uploadModel(ac, "../fixtures/test_model.csv"); err != nil {
		panic(err)
	}

	// test data
	if err := uploadData(ac, "../fixtures/test_data.csv"); err != nil {
		panic(err)
	}

	ac.Close()
}

func uploadModel(ac *db.AerospikeClient, testDataPath string) error {
	f, err := os.OpenFile(testDataPath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)

	i := 0
	for sc.Scan() {
		line := sc.Text() // GET the line string

		// skip header of csv file
		if i <= 0 {
			i++
			continue
		}

		splittedLine := strings.Split(line, ";")
		if len(splittedLine) != 4 {
			return errors.New("fixtures contains the wrong columns amount")
		}

		// order should be: publicationPoint;campaign;signalType;stage
		m, err := models.NewModel(splittedLine[0], splittedLine[1], splittedLine[2], ac)
		if err != nil {
			return err
		}

		// publish model
		if models.StageType(splittedLine[3]) == models.PUBLISHED {
			if err := m.PublishModel(ac); err != nil {
				return err
			}
		}

		i++
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return nil
}

func uploadData(ac *db.AerospikeClient, testDataPath string) error {
	f, err := os.OpenFile(testDataPath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)

	i := 0
	for sc.Scan() {
		line := sc.Text() // GET the line string

		// skip header of csv file
		if i <= 0 {
			i++
			continue
		}

		// order should be: model;signal;items
		splittedLine := strings.Split(line, ";")
		if len(splittedLine) != 3 {
			return errors.New("fixtures contains the wrong columns amount")
		}

		// need to break down the name of the model to fetch the actual object
		// from aerospike
		mn := strings.Split(splittedLine[0], "#")

		// TODO: this is slow! Need to find a smarter way. it works fine with small fixtures files
		m, err := models.GetExistingModel(mn[0], mn[1], ac)
		if err != nil {
			return err
		}

		sn := m.ComposeSetName()
		items := strings.Split(splittedLine[2], ",")
		if err := ac.AddOne(sn, splittedLine[1], splittedLine[1], items); err != nil {
			return err
		}

		i++
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return nil
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
