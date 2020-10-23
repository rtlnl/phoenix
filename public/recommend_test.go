package public

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/rtlnl/phoenix/models"

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

func TestRecommendCacheFlushing(t *testing.T) {
	t.Skip()

	dbc, c := GetTestRedisClient()
	defer c()

	// Test object creation
	if _, err := models.NewModel("cachemodel", "", []string{"signal"}, dbc); err != nil {
		t.Error("Failed to create model")
	}

	if _, err := models.NewContainer("cachepublication", "cachecampaign", []string{"cachemodel"}, dbc); err != nil {
		t.Error("Failed to create campaign")
	}

	UploadTestData(t, dbc, "testdata/test_published_model_data.jsonl", "cachemodel")

	code, body, err := MockRequest(http.MethodGet, "/v1/recommend?publicationPoint=cachepublication&campaign=cachecampaign&model=cachemodel&signalId=500083", nil)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "{\"modelName\":\"cachemodel\",\"recommendations\":[{\"item\":\"6456\",\"score\":\"0.6\"},{\"item\":\"1252\",\"score\":\"0.345\"},{\"item\":\"7876\",\"score\":\"0.987\"}]}", string(b))

	// delete recommendation from database
	err = dbc.DeleteOne("cachemodel", "500083")
	if err != nil {
		t.Error("Something went wrong deleting the entry")
		return
	}

	// do the same request again
	code, body, err = MockRequest(http.MethodGet, "/v1/recommend?publicationPoint=cachepublication&campaign=cachecampaign&model=cachemodel&signalId=500083", nil)
	if err != nil {
		t.Fail()
	}

	b, err = ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	// should still return same results even after removal from database because of caching
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "{\"modelName\":\"cachemodel\",\"recommendations\":[{\"item\":\"6456\",\"score\":\"0.6\"},{\"item\":\"1252\",\"score\":\"0.345\"},{\"item\":\"7876\",\"score\":\"0.987\"}]}", string(b))

	// now add `flushCache` param to the request
	code, body, err = MockRequest(http.MethodGet, "/v1/recommend?publicationPoint=cachepublication&campaign=cachecampaign&model=cachemodel&signalId=500083&flushCache=true", nil)
	if err != nil {
		t.Fail()
	}

	b, err = ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	// Now the cache should be updated and say the item is not found
	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"error\":\"key 500083 not found\"}", string(b))
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
	assert.Equal(t, msg, "{\"error\":\"Request format error: publicationPoint, campaign or signalId are missing\"}")
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
	assert.Equal(t, msg, "{\"error\":\"Request format error: publicationPoint, campaign or signalId are missing\"}")
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
	assert.Equal(t, msg, "{\"error\":\"Request format error: publicationPoint, campaign or signalId are missing\"}")
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
	assert.Equal(t, "{\"error\":\"Request format error: publicationPoint, campaign or signalId are missing\"}", msg)
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
	assert.Equal(t, "{\"error\":\"container with publication point tuna and campaign hello not found\"}", string(b))
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
	assert.Equal(t, "{\"error\":\"key jjkk_767 not found\"}", string(b))
}

func BenchmarkRecommend(b *testing.B) {
	b.StopTimer()

	// instantiate Redis client
	dbc, c := GetTestRedisClient()
	defer c()

	// Test object creation
	if _, err := models.NewModel("benchmark", "", []string{"signal"}, dbc); err != nil {
		b.FailNow()
	}

	if _, err := models.NewContainer("publication1", "campaign", []string{"benchmark"}, dbc); err != nil {
		b.FailNow()
	}

	// upload data to model
	UploadTestData(nil, dbc, "testdata/test_published_model_data.jsonl", "benchmark")

	for i := 0; i < b.N; i++ {
		MockRequestBenchmark(b, http.MethodGet, "/v1/recommend?publicationPoint=publication1&campaign=campaign&signalId=500083", nil)
	}
}
