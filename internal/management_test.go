package internal

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createManagementModelRequest(publicationPoint, campaign string, signalOrder []string, concatenator string) (*bytes.Reader, error) {
	mmr := &ManagementModelRequest{
		PublicationPoint: publicationPoint,
		Campaign:         campaign,
		SignalOrder:      signalOrder,
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
	truncate := CreateTestModel(t, ac, "rtl_nieuws", "homepage", "articleId", false)
	defer truncate()

	code, body, err := MockRequest(http.MethodGet, "/management/model?publicationPoint=rtl_nieuws&campaign=homepage", nil)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "{\"publicationPoint\":\"rtl_nieuws\",\"campaign\":\"homepage\",\"stage\":\"STAGED\",\"version\":\"0.1.0\",\"signalType\":\"articleId\"}", string(b))
}

func TestGetModelEmptyParams(t *testing.T) {
	code, body, err := MockRequest(http.MethodGet, "/management/model?&campaign=homepage", nil)
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
	code, body, err := MockRequest(http.MethodGet, "/management/model?publicationPoint=rtl_nieuws&campaign=panini", nil)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"message\":\"key panini does not exist\"}", string(b))
}

func TestCreateModelAlreadyExists(t *testing.T) {
	// get client
	ac, c := GetTestAerospikeClient()
	defer c()

	// create model
	truncate := CreateTestModel(t, ac, "kiwi", "oranges", "grapeId", false)
	defer truncate()

	r, err := createManagementModelRequest("kiwi", "oranges", []string{"grapeId"}, "")
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/management/model", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusUnprocessableEntity, code)
	assert.Equal(t, "{\"message\":\"model with publicationPoint 'kiwi' and campaign 'oranges' exists already\"}", string(b))
}

func TestCreateModelFailValidation(t *testing.T) {
	r, err := createManagementModelRequest("", "", nil, "_")
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/management/model", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	msg := string(b)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, true, strings.Contains(msg, "Error:Field validation for 'Campaign' failed on the 'required' tag"))
}

func TestEmptyModel(t *testing.T) {
	// get client
	ac, c := GetTestAerospikeClient()
	defer c()

	// create model
	truncate := CreateTestModel(t, ac, "banana", "pears", "appleId", false)
	defer truncate()

	r, err := createManagementModelRequest("banana", "pears", []string{"appleId"}, "")
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodDelete, "/management/model", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "{\"message\":\"model empty\"}", string(b))
}

func TestEmptyModelFailValidation(t *testing.T) {
	r, err := createManagementModelRequest("", "oranges", []string{""}, "_")
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodDelete, "/management/model", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	msg := string(b)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, true, strings.Contains(msg, "Error:Field validation for 'PublicationPoint' failed on the 'required' tag"))
}

func TestEmptyModelNotExist(t *testing.T) {
	r, err := createManagementModelRequest("pizza", "pepperoni", []string{"ham"}, "_")
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodDelete, "/management/model", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"message\":\"key pepperoni does not exist\"}", string(b))
}

func TestPublishModelAlreadyPublished(t *testing.T) {
	// get client
	ac, c := GetTestAerospikeClient()
	defer c()

	// create model
	truncate := CreateTestModel(t, ac, "kiwi", "oranges", "appleId", true)
	defer truncate()

	r, err := createManagementModelRequest("kiwi", "oranges", []string{"appleId"}, "")
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/management/model/publish", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "{\"message\":\"model is already published\"}", string(b))
}

func TestPublishModelFailValidation(t *testing.T) {
	r, err := createManagementModelRequest("", "oranges", []string{"grapeId"}, "_")
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/management/model/publish", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	msg := string(b)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, true, strings.Contains(msg, "Error:Field validation for 'PublicationPoint' failed on the 'required' tag"))
}

func TestPublishModelNotExist(t *testing.T) {
	r, err := createManagementModelRequest("salami", "pepperoni", []string{"pineappleId"}, "_")
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/management/model/publish", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"message\":\"key pepperoni does not exist\"}", string(b))
}

func TestConcatenatorFailValidation(t *testing.T) {
	r, err := createManagementModelRequest("salami", "pepperoni", []string{"pineappleId", "cheeseId"}, "+")
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/management/model", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "{\"message\":\"Key: 'ManagementModelRequest.Concatenator' Error:Field validation for 'Concatenator' failed on the 'contatenatorvalidator' tag\"}", string(b))
}

func TestConcatenatorPassValidation(t *testing.T) {
	r, err := createManagementModelRequest("kiwi", "oranges", []string{"appleId", "bananasId"}, "_")
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/management/model", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusCreated, code)
	assert.Equal(t, "{\"message\":\"model created\"}", string(b))
}

func TestMM(t *testing.T) {
	r, err := createManagementModelRequest("ham3", "pepperoni", []string{"pineappleId", "cheeseId"}, "")
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/management/model", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusCreated, code)
	assert.Equal(t, "{\"message\":\"model created\"}", string(b))
}

func TestMM1(t *testing.T) {
	r, err := createManagementModelRequest("ham3", "pepperoni", []string{"pineappleId"}, "")
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/management/model", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusCreated, code)
	assert.Equal(t, "{\"message\":\"model created\"}", string(b))
}
