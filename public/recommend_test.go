package public

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecommend(t *testing.T) {
	// get client
	ac, c := GetTestAerospikeClient()
	defer c()

	// create model
	truncate1 := CreateTestModel(t, ac, "rtl_nieuws", "homepage", "articleId", true)
	defer truncate1()

	truncate2 := UploadTestData(t, ac, "testdata/test_published_model_data.jsonl", "rtl_nieuws#homepage")
	defer truncate2()

	code, body, err := MockRequest(http.MethodGet, "/recommend?publicationPoint=rtl_nieuws&campaign=homepage&signalId=500083", nil)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "{\"recommendations\":[{\"item\":\"6456\",\"score\":\"0.6\"},{\"item\":\"1252\",\"score\":\"0.345\"},{\"item\":\"7876\",\"score\":\"0.987\"}]}", string(b))
}

func TestRecommendFailValidation(t *testing.T) {
	code, body, err := MockRequest(http.MethodGet, "/recommend?campaign=homepage&signalId=500083", nil)
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

func TestRecommendNoModel(t *testing.T) {
	code, body, err := MockRequest(http.MethodGet, "/recommend?publicationPoint=tuna&campaign=hello&signalId=500083", nil)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"message\":\"key hello does not exist\"}", string(b))
}

func TestRecommendWrongSignal(t *testing.T) {
	// get client
	ac, c := GetTestAerospikeClient()
	defer c()

	// create model
	truncate1 := CreateTestModel(t, ac, "rtl_nieuws", "homepage", "articleId", true)
	defer truncate1()

	code, body, err := MockRequest(http.MethodGet, "/recommend?publicationPoint=rtl_nieuws&campaign=homepage&signalId=jjkk_767", nil)
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

func TestRecommendModelStaged(t *testing.T) {
	// get client
	ac, c := GetTestAerospikeClient()
	defer c()

	// create model
	truncate1 := CreateTestModel(t, ac, "rtl_nieuws", "banana", "articleId", false)
	defer truncate1()

	code, body, err := MockRequest(http.MethodGet, "/recommend?publicationPoint=rtl_nieuws&campaign=banana&signalId=500083", nil)
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

	for i := 0; i < b.N; i++ {
		MockRequestBenchmark(b, http.MethodGet, "/recommend?publicationPoint=rtl_nieuws&campaign=homepage&signalId=500083", nil)
	}
}
