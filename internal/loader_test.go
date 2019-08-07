package internal

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createStreamingRequest(publicationPoint, campaign, signal string, recommendations []string) (*bytes.Reader, error) {
	rr := &StreamingRequest{
		PublicationPoint: publicationPoint,
		Campaign:         campaign,
		Signal:           signal,
		Recommendations:  recommendations,
	}

	rb, err := json.Marshal(rr)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(rb), nil
}

func createBatchRequestDirect(publicationPoint, campaign string, data []BatchData) (*bytes.Reader, error) {
	br := &BatchRequest{
		PublicationPoint: publicationPoint,
		Campaign:         campaign,
		Data:             data,
	}

	rb, err := json.Marshal(br)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(rb), nil
}

func createBatchRequestLocation(publicationPoint, campaign string, dataLocation string) (*bytes.Reader, error) {
	br := &BatchRequest{
		PublicationPoint: publicationPoint,
		Campaign:         campaign,
		DataLocation:     dataLocation,
	}

	rb, err := json.Marshal(br)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(rb), nil
}

func TestStreaming(t *testing.T) {
	signal := "100"
	recommendationItems := []string{"1", "2", "3", "4"}

	rb, err := createStreamingRequest("rtl_nieuws", "homepage", signal, recommendationItems)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/streaming", rb)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusCreated, code)
	assert.Equal(t, "{\"message\":\"signal 100 created\"}", string(b))
}

func TestStreamingModelNotExist(t *testing.T) {
	signal := "100"
	recommendationItems := []string{"1", "2", "3", "4"}

	rb, err := createStreamingRequest("pasta", "pizza", signal, recommendationItems)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/streaming", rb)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "{\"message\":\"key pizza does not exist\"}", string(b))
}

func TestStreamingUpdateData(t *testing.T) {
	signal := "100"
	recommendationItems := []string{"6", "7", "8", "9"}

	rb, err := createStreamingRequest("rtl_nieuws", "homepage", signal, recommendationItems)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPut, "/streaming", rb)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "{\"message\":\"signal 100 updated\"}", string(b))
}

func TestStreamingDeleteData(t *testing.T) {
	signal := "100"
	recommendationItems := []string{"6", "7", "8", "9"}

	rb, err := createStreamingRequest("rtl_nieuws", "homepage", signal, recommendationItems)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodDelete, "/streaming", rb)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "{\"message\":\"signal 100 deleted\"}", string(b))
}

func TestBatchUploadDirect(t *testing.T) {

}

func TestBatchUploadDirectModelPublished(t *testing.T) {

	bd := make([]BatchData, 1)
	d := []string{"item1", "item2", "item3"}
	bd[0] = map[string][]string{
		"123": d,
	}

	rb, err := createBatchRequestDirect("rtl_nieuws", "homepage", bd)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/batch", rb)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "{\"message\":\"you cannot upload data on already published models. stage it first\"}", string(b))
}

func TestBatchUploadDirectModelNotExist(t *testing.T) {

}

func TestBatchUploadS3(t *testing.T) {

}
