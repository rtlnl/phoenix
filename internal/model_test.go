package internal

import (
	"bytes"
	"fmt"
	"github.com/rtlnl/phoenix/models"
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
	dbc, c := GetTestRedisClient()
	defer c()

	// create model
	if _, err := models.NewModel("getter", "", []string{"articleId"}, dbc); err != nil {
		t.FailNow()
	}

	code, body, err := MockRequest(http.MethodGet, "/v1/management/models/?name=getter", nil)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "{\"model\":{\"name\":\"getter\",\"signalOrder\":[\"articleId\"],\"concatenator\":\"\"},\"message\":\"model fetched\"}", string(b))
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
	code, body, err := MockRequest(http.MethodGet, "/v1/management/models/?publicationPoint=rtl_nieuws&campaign=panini&name=ocean", nil)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"message\":\"model with name ocean not found\"}", string(b))
}

func TestCreateModelAlreadyExists(t *testing.T) {
	// get client
	dbc, c := GetTestRedisClient()
	defer c()

	// create model
	if _, err := models.NewModel("already", "", []string{"grapeId"}, dbc); err != nil {
		t.FailNow()
	}

	r, err := createManagementModelRequest("already", "", []string{"grapeId"})
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

	assert.Equal(t, http.StatusUnprocessableEntity, code)
	assert.Equal(t, "{\"message\":\"model with name already already exists\"}", string(b))
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
	dbc, c := GetTestRedisClient()
	defer c()

	// create model
	if _, err := models.NewModel("empty", "", []string{"appleId"}, dbc); err != nil {
		fmt.Print(err.Error())
		t.FailNow()
	}

	r, err := createManagementModelRequest("empty", "", []string{"appleId"})
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
	assert.Equal(t, "{\"model\":{\"name\":\"empty\",\"signalOrder\":[\"appleId\"],\"concatenator\":\"\"},\"message\":\"model empty\"}", string(b))
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
	assert.Equal(t, "{\"message\":\"model with name goat not found\"}", string(b))
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
	r, err := createManagementModelRequest("tech", "_", []string{"office", "floor"})
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
	assert.Equal(t, "{\"model\":{\"name\":\"tech\",\"signalOrder\":[\"office\",\"floor\"],\"concatenator\":\"_\"},\"message\":\"model created\"}", string(b))
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

func TestGetDataPreview(t *testing.T) {
	// get client
	dbc, c := GetTestRedisClient()
	defer c()

	//create model
	if _, err := models.NewModel("preview", "", []string{"userId"}, dbc); err != nil {
		fmt.Print(err.Error())
		t.FailNow()
	}

	bd := make([]BatchData, 1)
	d := []models.ItemScore{
		{
			"item":  "111",
			"score": "0.6",
			"type":  "movie",
		},
		{
			"item":  "222",
			"score": "0.4",
			"type":  "movie",
		},
		{
			"item":  "555",
			"score": "0.16",
			"type":  "series",
		},
	}
	bd[0] = map[string][]models.ItemScore{
		"123": d,
	}

	rb, err := createBatchRequestDirect("preview", bd)
	if err != nil {
		t.Fail()
	}

	// upload data
	status, _, err := MockRequest(http.MethodPost, "/v1/batch", rb)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusCreated, status)

	code, body, err := MockRequest(http.MethodGet, "/v1/management/models/preview?name=preview", nil)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t,"{\"preview\":[{\"signalId\":\"123\",\"recommended\":[{\"item\":\"111\",\"score\":\"0.6\",\"type\":\"movie\"},{\"item\":\"222\",\"score\":\"0.4\",\"type\":\"movie\"},{\"item\":\"555\",\"score\":\"0.16\",\"type\":\"series\"}]}]}", string(b))
}
