package public

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createRecommendRequest(publicationPoint, campaign string, signals []Signal) (*bytes.Reader, error) {
	rr := &RecommendRequest{
		PublicationPoint: publicationPoint,
		Campaign:         campaign,
		Signals:          signals,
	}

	rb, err := json.Marshal(rr)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(rb), nil
}

func TestRecommend(t *testing.T) {
	// get client
	ac, c := GetAerospikeClient()
	defer c()

	// create model
	truncate1 := CreateTestModel(t, ac, "rtl_nieuws", "homepage", "articleId", true)
	defer truncate1()

	truncate2 := UploadTestData(t, ac, "testdata/test_published_model_data.jsonl", "rtl_nieuws#homepage")
	defer truncate2()

	ss := make([]Signal, 1)
	ss[0] = Signal{
		"articleId": "500083",
	}

	rb, err := createRecommendRequest("rtl_nieuws", "homepage", ss)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/recommend", rb)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "{\"signals\":[{\"articleId\":\"500083\"}],\"recommendations\":[{\"item\":\"6456\",\"score\":\"0.6\"},{\"item\":\"1252\",\"score\":\"0.345\"},{\"item\":\"7876\",\"score\":\"0.987\"}]}", string(b))
}

func TestRecommendFailValidation(t *testing.T) {
	rb, err := createRecommendRequest("", "", nil)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/recommend", rb)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	msg := string(b)

	assert.Equal(t, http.StatusBadRequest, code)

	// cannot check the message verbatim because gin could place the validation error
	// in a different position
	assert.Equal(t, true, strings.Contains(msg, "'RecommendRequest.PublicationPoint' Error:Field validation for 'PublicationPoint' failed on the 'required' tag"))
}

func TestRecommendNoModel(t *testing.T) {
	ss := make([]Signal, 2)
	rb, err := createRecommendRequest("chicken", "tuna", ss)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/recommend", rb)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"message\":\"key tuna does not exist\"}", string(b))
}

func TestRecommendWrongSignal(t *testing.T) {
	// get client
	ac, c := GetAerospikeClient()
	defer c()

	// create model
	truncate1 := CreateTestModel(t, ac, "rtl_nieuws", "homepage", "articleId", true)
	defer truncate1()

	ss := make([]Signal, 1)
	ss[0] = Signal{
		"articleId_sloths": "500083",
	}

	rb, err := createRecommendRequest("rtl_nieuws", "homepage", ss)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/recommend", rb)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "{\"message\":\"signal is not formatted correctly\"}", string(b))
}

func TestRecommendModelStaged(t *testing.T) {
	// get client
	ac, c := GetAerospikeClient()
	defer c()

	// create model
	truncate1 := CreateTestModel(t, ac, "rtl_nieuws", "banana", "articleId", false)
	defer truncate1()

	ss := make([]Signal, 1)
	ss[0] = Signal{
		"appleId": "500083",
	}

	rb, err := createRecommendRequest("rtl_nieuws", "banana", ss)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/recommend", rb)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "{\"message\":\"model is staged. Clients cannot access staged models\"}", string(b))
}

func BenchmarkRecommend(b *testing.B) {
	b.StopTimer()

	ss := make([]Signal, 1)
	ss[0] = Signal{
		"articleId": "500083",
	}

	rb, err := createRecommendRequest("rtl_nieuws", "homepage", ss)
	if err != nil {
		b.Fail()
	}

	for i := 0; i < b.N; i++ {
		MockRequestBenchmark(b, http.MethodPost, "/recommend", rb)
	}
}
