package public

import (
	"github.com/rtlnl/phoenix/models"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecommend(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	// Test object creation
	if _, err := models.NewModel("model", "", []string{"signal"}, dbc); err != nil {
		t.FailNow()
	}

	if _, err := models.NewContainer("publication", "campaign", []string{"model"}, dbc); err != nil {
		t.FailNow()
	}

	UploadTestData(t, dbc, "testdata/test_published_model_data.jsonl", "model")

	code, body, err := MockRequest(http.MethodGet, "/v1/recommend?publicationPoint=publication&campaign=campaign&model=model&signalId=500083", nil)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "{\"modelName\":\"model\",\"recommendations\":[{\"item\":\"6456\",\"score\":\"0.6\"},{\"item\":\"1252\",\"score\":\"0.345\"},{\"item\":\"7876\",\"score\":\"0.987\"}]}", string(b))
}

func TestRecommendFailValidation1(t *testing.T) {
	code, body, err := MockRequest(http.MethodGet, "/v1/recommend?campaign=homepage&signalId=500083", nil)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	msg := string(b)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, msg, "{\"message\":\"missing publicationPoint in the URL query\"}")
}

func TestRecommendFailValidation2(t *testing.T) {
	code, body, err := MockRequest(http.MethodGet, "/v1/recommend?publicationPoint=hello&signalId=500083", nil)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	msg := string(b)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, msg, "{\"message\":\"missing campaign in the URL query\"}")
}

func TestRecommendFailValidation3(t *testing.T) {
	code, body, err := MockRequest(http.MethodGet, "/v1/recommend?publicationPoint=hello&campaign=homepage", nil)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	msg := string(b)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, msg, "{\"message\":\"missing signalId in the URL query\"}")
}

func TestRecommendFailValidation4(t *testing.T) {
	code, body, err := MockRequest(http.MethodGet, "/v1/recommend?campaign=homepage", nil)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	msg := string(b)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "{\"message\":\"missing publicationPoint,signalId in the URL query\"}", msg)
}

func TestRecommendNoModel(t *testing.T) {
	code, body, err := MockRequest(http.MethodGet, "/v1/recommend?publicationPoint=tuna&campaign=hello&model=banana&signalId=500083", nil)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"message\":\"container with publication point tuna and campaign hello not found\"}", string(b))
}

func TestRecommendWrongSignal(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	// Test object creation
	if _, err := models.NewModel("wrongsig", "", []string{"signal"}, dbc); err != nil {
		t.FailNow()
	}

	if _, err := models.NewContainer("wrong", "campaign", []string{"wrongsig"}, dbc); err != nil {
		t.FailNow()
	}

	UploadTestData(t, dbc, "testdata/test_published_model_data.jsonl", "wrongsig")

	code, body, err := MockRequest(http.MethodGet, "/v1/recommend?publicationPoint=wrong&campaign=campaign&model=wrongsig&signalId=jjkk_767", nil)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"message\":\"key jjkk_767 not found\"}", string(b))
}

func BenchmarkRecommend(b *testing.B) {
	b.StopTimer()

	// instantiate Redis client
	dbc, c := GetTestRedisClient()
	defer c()

	// Test object creation
	if _, err := models.NewModel("model", "", []string{"signal"}, dbc); err != nil {
		b.FailNow()
	}

	if _, err := models.NewContainer("publication1", "campaign", []string{"model"}, dbc); err != nil {
		b.FailNow()
	}

	// upload data to model
	UploadTestData(nil, dbc, "testdata/test_published_model_data.jsonl", "model")

	for i := 0; i < b.N; i++ {
		MockRequestBenchmark(b, http.MethodGet, "/v1/recommend?publicationPoint=publication1&campaign=campaign&signalId=500083", nil)
	}
}
