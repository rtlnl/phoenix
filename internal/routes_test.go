package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	aero "github.com/aerospike/aerospike-client-go"
	"github.com/gin-gonic/gin"
	"github.com/rtlnl/data-personalization-api/middleware"
	"github.com/rtlnl/data-personalization-api/pkg/db"
)

const (
	testDBHost    = "127.0.0.1"
	testDBPort    = 3000
	testNamespace = "test"
	testSet       = "test_model"
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
}

func tearDown() {
	router = nil
}

func loadFixtures() {
	ac := db.NewAerospikeClient(testDBHost, testNamespace, testDBPort)

	// truncate sets first to clean db for security
	if err := ac.TruncateSet(testSet); err != nil {
		panic(err)
	}

	// load fixtures here
	if err := uploadData(ac, "../fixtures/test_data.csv"); err != nil {
		panic(err)
	}

	ac.Close()
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

		// generate bin and append to array
		k, b := createBin(line, fmt.Sprintf("item_%d", i))
		if k == nil || b == nil {
			continue
		}

		// store record
		if err := ac.Client.PutBins(ac.Client.DefaultWritePolicy, k, b); err != nil {
			panic(err)
		}
		i++
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return nil
}

func createBin(line, binKey string) (*aero.Key, *aero.Bin) {
	splittedLine := strings.Split(line, ";")
	if len(splittedLine) != 2 {
		return nil, nil
	}

	key, err := strconv.Unquote(splittedLine[0])
	if err != nil {
		return nil, nil
	}

	items := strings.Split(splittedLine[1], ",")

	ak, err := aero.NewKey(testNamespace, testSet, key)
	if err != nil {
		return nil, nil
	}

	return ak, aero.NewBin(binKey, items)
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
