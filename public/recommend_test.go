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
	truncateModel := CreateTestModel(t, ac, "collaborative", "", []string{"articleId"}, true)
	defer truncateModel()

	// create container
	truncateContainer := CreateTestContainer(t, ac, "rtl_nieuws", "homepage", []string{"collaborative"})
	defer truncateContainer()

	truncateTestData := UploadTestData(t, ac, "testdata/test_published_model_data.jsonl", "collaborative")
	defer truncateTestData()

	code, body, err := MockRequest(http.MethodGet, "/recommend?publicationPoint=rtl_nieuws&campaign=homepage&model=collaborative&signalId=500083", nil)
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

func TestRecommendFailValidation1(t *testing.T) {
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

func TestRecommendFailValidation2(t *testing.T) {
	code, body, err := MockRequest(http.MethodGet, "/recommend?publicationPoint=hello&signalId=500083", nil)
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
	code, body, err := MockRequest(http.MethodGet, "/recommend?publicationPoint=hello&campaign=homepage", nil)
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
	code, body, err := MockRequest(http.MethodGet, "/recommend?campaign=homepage", nil)
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
	code, body, err := MockRequest(http.MethodGet, "/recommend?publicationPoint=tuna&campaign=hello&model=banana&signalId=500083", nil)
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
	// get client
	ac, c := GetTestAerospikeClient()
	defer c()

	// create model
	truncateModel := CreateTestModel(t, ac, "ciao", "", []string{"articleId"}, true)
	defer truncateModel()

	// create container
	truncateContainer := CreateTestContainer(t, ac, "curry", "homepage", []string{"ciao"})
	defer truncateContainer()

	code, body, err := MockRequest(http.MethodGet, "/recommend?publicationPoint=curry&campaign=homepage&model=ciao&signalId=jjkk_767", nil)
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
	truncateModel := CreateTestModel(t, ac, "sloth", "", []string{"articleId"}, false)
	defer truncateModel()

	// create container
	truncateContainer := CreateTestContainer(t, ac, "fruits", "banana", []string{"sloth"})
	defer truncateContainer()

	code, body, err := MockRequest(http.MethodGet, "/recommend?publicationPoint=fruits&campaign=banana&model=sloth&signalId=500083", nil)
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

	// get client
	ac, c := GetTestAerospikeClient()
	defer c()

	// create model
	truncateModel := CreateTestModel(nil, ac, "collaborative", "", []string{"articleId"}, false)
	defer truncateModel()

	// create container
	truncateContainer := CreateTestContainer(nil, ac, "rtl_nieuws", "homepage", []string{"collaborative"})
	defer truncateContainer()

	// upload data to model
	truncateTestData := UploadTestData(nil, ac, "testdata/test_published_model_data.jsonl", "collaborative")
	defer truncateTestData()

	for i := 0; i < b.N; i++ {
		MockRequestBenchmark(b, http.MethodGet, "/recommend?publicationPoint=rtl_nieuws&campaign=homepage&signalId=500083", nil)
	}
}
