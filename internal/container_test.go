package internal

import (
	"bytes"
	"github.com/rtlnl/phoenix/models"
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
	// instantiate Redis client
	dbc, c := GetTestRedisClient()
	defer c()

	// Test object creation
	if _, err := models.NewModel("quattro-formaggi", "", []string{"gorgonzola"}, dbc); err != nil {
		t.FailNow()
	}

	if _, err := models.NewContainer("food", "pizza", []string{"quattro-formaggi"}, dbc); err != nil {
		t.FailNow()
	}

	code, body, err := MockRequest(http.MethodGet, "/v1/management/containers/?publicationPoint=food&campaign=pizza", nil)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "{\"container\":{\"publicationPoint\":\"food\",\"campaign\":\"pizza\",\"models\":[\"quattro-formaggi\"]},\"message\":\"container fetched\"}", string(b))
}

func TestGetContainerEmptyParams(t *testing.T) {
	// instantiate Redis client
	//dbc, c := GetTestRedisClient()
	//defer c()
	//
	//if _, err := models.NewContainer("", "", []string{"collaborative"}, dbc); err != nil {
	//	t.FailNow()
	//}

	code, body, err := MockRequest(http.MethodGet, "/v1/management/containers/?campaign=homepage", nil)
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
	code, body, err := MockRequest(http.MethodGet, "/v1/management/containers/?publicationPoint=ciao&campaign=panini", nil)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"message\":\"container with publication point ciao and campaign panini not found\"}", string(b))
}

func TestCreateContainerAlreadyExists(t *testing.T) {
	// instantiate Redis client
	dbc, c := GetTestRedisClient()
	defer c()

	// Test object creation
	if _, err := models.NewModel("animals", "", []string{"paw"}, dbc); err != nil {
		t.FailNow()
	}

	if _, err := models.NewContainer("dog", "vizsla", []string{"animals"}, dbc); err != nil {
		t.FailNow()
	}

	r, err := createManagementContainerRequest("dog", "vizsla", []string{"animals"})
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/v1/management/containers/", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusUnprocessableEntity, code)
	assert.Equal(t, "{\"message\":\"container with publication point dog and campaign vizsla already exists\"}", string(b))
}

func TestCreateContainerFailValidationCampaign(t *testing.T) {
	r, err := createManagementContainerRequest("videoland", "", nil)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/v1/management/containers/", r)
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

	code, body, err := MockRequest(http.MethodPost, "/v1/management/containers/", r)
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

	code, body, err := MockRequest(http.MethodPost, "/v1/management/containers/", r)
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
	dbc, c := GetTestRedisClient()
	defer c()

	// Test object creation
	if _, err := models.NewModel("egypt", "", []string{"god"}, dbc); err != nil {
		t.FailNow()
	}

	if _, err := models.NewContainer("cat", "anubi", []string{"egypt"}, dbc); err != nil {
		t.FailNow()
	}

	r, err := createManagementContainerRequest("cat", "anubi", []string{"egypt"})
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodDelete, "/v1/management/containers/", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "{\"container\":{\"publicationPoint\":\"cat\",\"campaign\":\"anubi\",\"models\":[\"egypt\"]},\"message\":\"container empty\"}", string(b))
}

func TestEmptyContainerFailValidation(t *testing.T) {
	r, err := createManagementContainerRequest("", "profile", []string{""})
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodDelete, "/v1/management/containers/", r)
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
	r, err := createManagementContainerRequest("rtl", "clock", []string{"test"})
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodDelete, "/v1/management/containers/", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"message\":\"container with publication point rtl and campaign clock not found\"}", string(b))
}

func TestLinkModel(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	r, err := createManagementContainerRequest("channel", "dart", []string{"hello", "world"})
	if err != nil {
		t.Fail()
	}

	// Test object creation
	if _, err := models.NewModel("hello", "", []string{"articleId"}, dbc); err != nil {
		t.FailNow()
	}

	// Test object creation
	if _, err := models.NewModel("world", "", []string{"articleId"}, dbc); err != nil {
		t.FailNow()
	}

	if _, err = models.NewContainer("channel", "dart", []string{""}, dbc); err != nil {
		t.FailNow()
	}

	code, body, err := MockRequest(http.MethodPut, "/v1/management/containers/link-model", r)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "{\"container\":{\"publicationPoint\":\"channel\",\"campaign\":\"dart\",\"models\":[\"hello\",\"world\"]},\"message\":\"model linked to container\"}", string(b))
}
