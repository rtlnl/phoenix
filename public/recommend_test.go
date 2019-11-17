package public

import (
	"github.com/rtlnl/phoenix/models"
	"github.com/rtlnl/phoenix/pkg/db"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecommend(t *testing.T) {
	// instantiate Redis client
	redisClient, err := db.NewRedisClient(testDBHost, nil)
	if err != nil {
		panic(err)
	}
	defer redisClient.Close()

	// Test object creation
	if _, err := models.NewModel("model", "", []string{"signal"}, redisClient); err != nil {
		t.FailNow()
	}

	_, err = models.NewContainer("publication", "campaign", []string{"model"}, redisClient)
	if err != nil {
		t.FailNow()
	}

	UploadTestData(t, redisClient, "testdata/test_published_model_data.jsonl", "model")

	code, body, err := MockRequest(http.MethodGet, "/v1/recommend?publicationPoint=rtl_nieuws&campaign=homepage&model=collaborative&signalId=500083", nil)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "{\"modelName\":\"collaborative\",\"recommendations\":[{\"item\":\"6456\",\"score\":\"0.6\"},{\"item\":\"1252\",\"score\":\"0.345\"},{\"item\":\"7876\",\"score\":\"0.987\"}]}", string(b))
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
	assert.Equal(t, "{\"message\":\"container with publication point tuna and campaign hello is not found\"}", string(b))
}

func TestRecommendWrongSignal(t *testing.T) {
	// instantiate Redis client
	redisClient, err := db.NewRedisClient(testDBHost, nil)
	if err != nil {
		panic(err)
	}
	defer redisClient.Close()

	// Test object creation
	if _, err := models.NewModel("model", "", []string{"signal"}, redisClient); err != nil {
		t.FailNow()
	}

	_, err = models.NewContainer("publication1", "campaign", []string{"model"}, redisClient)
	if err != nil {
		t.FailNow()
	}

	code, body, err := MockRequest(http.MethodGet, "/v1/recommend?publicationPoint=curry&campaign=homepage&model=ciao&signalId=jjkk_767", nil)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"message\":\"key jjkk_767 does not exist\"}", string(b))
}

func BenchmarkRecommend(b *testing.B) {
	b.StopTimer()

	// instantiate Redis client
	redisClient, err := db.NewRedisClient(testDBHost, nil)
	if err != nil {
		panic(err)
	}
	defer redisClient.Close()

	// Test object creation
	if _, err := models.NewModel("model", "", []string{"signal"}, redisClient); err != nil {
		b.FailNow()
	}

	_, err = models.NewContainer("publication1", "campaign", []string{"model"}, redisClient)
	if err != nil {
		b.FailNow()
	}

	// upload data to model
	UploadTestData(nil, redisClient, "testdata/test_published_model_data.jsonl", "model")

	for i := 0; i < b.N; i++ {
		MockRequestBenchmark(b, http.MethodGet, "/v1/recommend?publicationPoint=rtl_nieuws&campaign=homepage&signalId=500083", nil)
	}
}
