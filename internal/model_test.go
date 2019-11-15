package internal

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createManagementModelRequest(name, concatenator string, signalOrder []string) (*bytes.Reader, error) {
	mmr := &ManagementModelRequest{
		Name:        name,
		SignalOrder: signalOrder,
	}

	if concatenator != "" {
		mmr.Concatenator = concatenator
	}

	rb, err := json.Marshal(mmr)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(rb), nil
}

func TestGetModel(t *testing.T) {
	// get client
	ac, c := GetTestAerospikeClient()
	defer c()

	// create model
	truncate := CreateTestModel(t, ac, "collaborative", "", []string{"articleId"}, false)
	defer truncate()

	code, body, err := MockRequest(http.MethodGet, "/v1/management/models/?name=collaborative", nil)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "{\"model\":{\"name\":\"collaborative\",\"stage\":\"STAGED\",\"version\":\"0.1.0\",\"signalOrder\":[\"articleId\"],\"concatenator\":\"\"},\"message\":\"model fetched\"}", string(b))
}

func TestGetModelEmptyParams(t *testing.T) {
	code, body, err := MockRequest(http.MethodGet, "/v1/management/models/?&campaign=homepage", nil)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "{\"message\":\"missing parameters in url for searching the model\"}", string(b))
}

func TestGetModelNotExist(t *testing.T) {
	code, body, err := MockRequest(http.MethodGet, "/v1/management/models/?publicationPoint=rtl_nieuws&campaign=panini&name=collaborative", nil)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"message\":\"model collaborative not found\"}", string(b))
}

func TestCreateModelAlreadyExists(t *testing.T) {
	// get client
	ac, c := GetTestAerospikeClient()
	defer c()

	// create model
	truncate := CreateTestModel(t, ac, "collaborative", "", []string{"grapeId"}, false)
	defer truncate()

	r, err := createManagementModelRequest("collaborative", "", []string{"grapeId"})
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/v1/management/models/", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusCreated, code)
	assert.Equal(t, "{\"model\":{\"name\":\"collaborative\",\"stage\":\"STAGED\",\"version\":\"0.1.0\",\"signalOrder\":[\"grapeId\"],\"concatenator\":\"\"},\"message\":\"model created\"}", string(b))
}

func TestCreateModelFailValidation(t *testing.T) {
	r, err := createManagementModelRequest("", "", nil)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/v1/management/models/", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	msg := string(b)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, true, strings.Contains(msg, "Error:Field validation for 'Name' failed on the 'required' tag"))
}

func TestEmptyModel(t *testing.T) {
	// get client
	ac, c := GetTestAerospikeClient()
	defer c()

	// create model
	truncate := CreateTestModel(t, ac, "collaborative", "", []string{"appleId"}, false)
	defer truncate()

	r, err := createManagementModelRequest("collaborative", "", []string{"appleId"})
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodDelete, "/v1/management/models/", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "{\"model\":{\"name\":\"collaborative\",\"stage\":\"STAGED\",\"version\":\"0.1.0\",\"signalOrder\":[\"appleId\"],\"concatenator\":\"\"},\"message\":\"model empty\"}", string(b))
}

func TestEmptyModelNotExist(t *testing.T) {
	r, err := createManagementModelRequest("goat", "", []string{"ham"})
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodDelete, "/v1/management/models/", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"message\":\"model goat not found\"}", string(b))
}

func TestPublishModelAlreadyPublished(t *testing.T) {
	// get client
	ac, c := GetTestAerospikeClient()
	defer c()

	// create model
	truncate := CreateTestModel(t, ac, "pears", "", []string{"appleId"}, true)
	defer truncate()

	r, err := createManagementModelRequest("pears", "", []string{"appleId"})
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/v1/management/models/publish", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "{\"message\":\"model is already PUBLISHED\"}", string(b))
}

func TestPublishModelFailValidation(t *testing.T) {
	r, err := createManagementModelRequest("", "", []string{"grapeId"})
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/v1/management/models/publish", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	msg := string(b)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, true, strings.Contains(msg, "Error:Field validation for 'Name' failed on the 'required' tag"))
}

func TestPublishModelNotExist(t *testing.T) {
	r, err := createManagementModelRequest("bear", "", []string{"pineappleId"})
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/v1/management/models/publish", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"message\":\"model bear not found\"}", string(b))
}

func TestStageModelNotExist(t *testing.T) {
	r, err := createManagementModelRequest("bear", "", []string{"pineappleId"})
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/v1/management/models/stage", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"message\":\"model bear not found\"}", string(b))
}

func TestStageModelAlreadyStaged(t *testing.T) {
	// get client
	ac, c := GetTestAerospikeClient()
	defer c()

	// create model
	truncate := CreateTestModel(t, ac, "grapes", "", []string{"appleId"}, false)
	defer truncate()

	r, err := createManagementModelRequest("grapes", "", []string{"appleId"})
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/v1/management/models/stage", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "{\"message\":\"model is already STAGED\"}", string(b))
}

func TestConcatenatorFailValidation(t *testing.T) {
	r, err := createManagementModelRequest("collaborative", "+", []string{"pineappleId", "cheeseId"})
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/v1/management/models/", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "{\"message\":\"for two or more signalOrder, a concatenator character from this list is mandatory: ["+strings.Join(concatenatorList, ", ")+"]\"}", string(b))
}

func TestConcatenatorPassValidation(t *testing.T) {
	r, err := createManagementModelRequest("collaborative", "_", []string{"appleId", "bananasId"})
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/v1/management/models/", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusCreated, code)
	assert.Equal(t, "{\"model\":{\"name\":\"collaborative\",\"stage\":\"STAGED\",\"version\":\"0.1.0\",\"signalOrder\":[\"appleId\",\"bananasId\"],\"concatenator\":\"_\"},\"message\":\"model created\"}", string(b))
}

func TestConcatenatorMissing(t *testing.T) {
	r, err := createManagementModelRequest("collaborative", "", []string{"pineappleId", "cheeseId"})
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/v1/management/models/", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "{\"message\":\"for two or more signalOrder, a concatenator character from this list is mandatory: ["+strings.Join(concatenatorList, ", ")+"]\"}", string(b))
}
func TestConcatenatorUneeded(t *testing.T) {
	r, err := createManagementModelRequest("collaborative", "0", []string{"pineappleId"})
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/v1/management/models/", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "{\"message\":\"for one signalOrder no concatenator character is required\"}", string(b))
}

func TestRemoveModelFromContainer(t *testing.T) {
	// get client
	ac, c := GetTestAerospikeClient()
	defer c()

	// create model hello
	truncateHello := CreateTestModel(t, ac, "hello", "", []string{"articleId"}, false)
	defer truncateHello()

	// create model world
	truncateWorld := CreateTestModel(t, ac, "world", "", []string{"articleId"}, false)
	defer truncateWorld()

	// create container
	truncate := CreateTestContainer(t, ac, "channel", "dart", []string{"hello", "world"})
	defer truncate()

	r, err := createManagementModelRequest("hello", "", []string{"articleId"})
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodDelete, "/v1/management/models/", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "{\"model\":{\"name\":\"hello\",\"stage\":\"STAGED\",\"version\":\"0.1.0\",\"signalOrder\":[\"articleId\"],\"concatenator\":\"\"},\"message\":\"model empty\"}", string(b))

	code, body, err = MockRequest(http.MethodGet, "/v1/management/containers/?publicationPoint=channel&campaign=dart", nil)
	if err != nil {
		t.Fail()
	}

	b, err = ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "{\"container\":{\"publicationPoint\":\"channel\",\"campaign\":\"dart\",\"models\":[\"world\"]},\"message\":\"container fetched\"}", string(b))
}
