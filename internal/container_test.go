package internal

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createManagementContainerRequest(publicationPoint, campaign string, models []string) (*bytes.Reader, error) {
	mmc := &ManagementContainerRequest{
		PublicationPoint: publicationPoint,
		Campaign:         campaign,
		Models:           models,
	}

	rb, err := json.Marshal(mmc)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(rb), nil
}

func TestGetContainer(t *testing.T) {
	// get client
	ac, c := GetTestAerospikeClient()
	defer c()

	// create model
	truncate := CreateTestContainer(t, ac, "videoland", "homepage", []string{"collaborative"})
	defer truncate()

	code, body, err := MockRequest(http.MethodGet, "/management/containers/?publicationPoint=videoland&campaign=homepage", nil)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "{\"container\":{\"publicationPoint\":\"videoland\",\"campaign\":\"homepage\",\"models\":[\"collaborative\"],\"createdAt\":null},\"message\":\"container fetched\"}", string(b))
}

func TestGetContainerEmptyParams(t *testing.T) {
	// get client
	ac, c := GetTestAerospikeClient()
	defer c()

	// create model
	truncate := CreateTestContainer(t, ac, "videoland", "homepage", []string{"collaborative"})
	defer truncate()

	code, body, err := MockRequest(http.MethodGet, "/management/containers/?campaign=homepage", nil)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "{\"message\":\"missing parameters in url for searching the container\"}", string(b))
}

func TestGetContainerNotExist(t *testing.T) {
	code, body, err := MockRequest(http.MethodGet, "/management/containers/?publicationPoint=goat&campaign=panini", nil)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"message\":\"container with publication point goat and campaign panini not found\"}", string(b))
}

func TestCreateContainerAlreadyExists(t *testing.T) {
	// get client
	ac, c := GetTestAerospikeClient()
	defer c()

	// create model
	truncate := CreateTestContainer(t, ac, "videoland", "homepage", []string{"collaborative"})
	defer truncate()

	r, err := createManagementContainerRequest("videoland", "homepage", []string{"collaborative"})
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/management/containers/", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusCreated, code)
	assert.Equal(t, "{\"container\":{\"publicationPoint\":\"videoland\",\"campaign\":\"homepage\",\"models\":[\"collaborative\"],\"createdAt\":null},\"message\":\"container created\"}", string(b))
}

func TestCreateContainerFailValidationCampaign(t *testing.T) {
	r, err := createManagementContainerRequest("videoland", "", nil)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/management/containers/", r)
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

func TestCreateContainerFailValidationPP(t *testing.T) {
	r, err := createManagementContainerRequest("", "homepage", nil)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/management/containers/", r)
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

func TestCreateContainerFailValidation(t *testing.T) {
	r, err := createManagementContainerRequest("", "", nil)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/management/containers/", r)
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

func TestEmptyContainer(t *testing.T) {
	// get client
	ac, c := GetTestAerospikeClient()
	defer c()

	// create container
	truncate := CreateTestContainer(t, ac, "videoland", "profile", []string{"test"})
	defer truncate()

	r, err := createManagementContainerRequest("videoland", "profile", []string{"test"})
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodDelete, "/management/containers/", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "{\"container\":{\"publicationPoint\":\"videoland\",\"campaign\":\"profile\",\"models\":null,\"createdAt\":null},\"message\":\"container empty\"}", string(b))
}

func TestEmptyContainerFailValidation(t *testing.T) {
	r, err := createManagementContainerRequest("", "profile", []string{""})
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodDelete, "/management/containers/", r)
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

func TestEmptyContainerNotExist(t *testing.T) {
	r, err := createManagementContainerRequest("rtl_news", "sport", []string{"test"})
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodDelete, "/management/containers/", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"message\":\"container with publication point rtl_news and campaign sport not found\"}", string(b))
}

func TestLinkModel(t *testing.T) {
	// get client
	ac, c := GetTestAerospikeClient()
	defer c()

	r, err := createManagementContainerRequest("channel", "dart", []string{"hello", "world"})
	if err != nil {
		t.Fail()
	}

	// create container
	CreateTestContainer(t, ac, "channel", "dart", []string{""})

	code, body, err := MockRequest(http.MethodPut, "/management/containers/link-model", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "{\"container\":{\"publicationPoint\":\"channel\",\"campaign\":\"dart\",\"models\":[\"hello\",\"world\"],\"createdAt\":null},\"message\":\"model linked to container\"}", string(b))

	// force truncate
	ac.TruncateSet("channel")
}
