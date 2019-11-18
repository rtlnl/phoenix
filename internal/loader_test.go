package internal

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	paws "github.com/rtlnl/phoenix/pkg/aws"

	"github.com/rtlnl/phoenix/models"
	"github.com/rtlnl/phoenix/pkg/db"
	"github.com/stretchr/testify/assert"
)

func createStreamingRequest(modelName, signal string, recommendations []models.ItemScore) (*bytes.Reader, error) {
	rr := &StreamingRequest{
		Signal:          signal,
		ModelName:       modelName,
		Recommendations: recommendations,
	}

	rb, err := json.Marshal(rr)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(rb), nil
}

func createBatchRequestDirect(modelName string, data []BatchData) (*bytes.Reader, error) {
	br := &BatchRequest{
		ModelName: modelName,
		Data:      data,
	}

	rb, err := json.Marshal(br)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(rb), nil
}

func createBatchRequestLocation(modelName string, dataLocation string) (*bytes.Reader, error) {
	br := &BatchRequest{
		ModelName:    modelName,
		DataLocation: dataLocation,
	}

	rb, err := json.Marshal(br)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(rb), nil
}

func TestStreaming(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	if _, err := models.NewModel("streaming", "", []string{"articleId"}, dbc); err != nil {
		t.FailNow()
	}

	signal := "100"
	recommendationItems := []models.ItemScore{
		{
			"item":  "111",
			"score": "0.6",
			"type":  "movie",
		},
		{
			"item":  "222",
			"score": "0.4",
			"type":  "movie",
		},
		{
			"item":  "555",
			"score": "0.16",
			"type":  "series",
		},
	}

	rb, err := createStreamingRequest("streaming", signal, recommendationItems)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/v1/streaming", rb)
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

func TestStreamingBadSignal(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	if _, err := models.NewModel("hybrid", "_", []string{"articleId","userId"}, dbc); err != nil {
		t.FailNow()
	}

	signal := "100"
	recommendationItems := []models.ItemScore{
		{
			"item":  "111",
			"score": "0.6",
			"type":  "movie",
		},
		{
			"item":  "222",
			"score": "0.4",
			"type":  "movie",
		},
		{
			"item":  "555",
			"score": "0.16",
			"type":  "series",
		},
	}

	rb, err := createStreamingRequest("hybrid", signal, recommendationItems)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/v1/streaming", rb)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "{\"message\":\"the expected signal format must be articleId_userId\"}", string(b))
}

func TestStreamingBadPayload(t *testing.T) {
	signal := ""
	recommendationItems := []models.ItemScore{}

	rb, err := createStreamingRequest("collaborative", signal, recommendationItems)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/v1/streaming", rb)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	msg := string(b)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, true, strings.Contains(msg, "'StreamingRequest.Signal' Error:Field validation for 'Signal' failed on the 'required' tag"))
}

func TestStreamingUpdateBadPayload(t *testing.T) {
	signal := ""
	recommendationItems := []models.ItemScore{}

	rb, err := createStreamingRequest("collaborative", signal, recommendationItems)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPut, "/v1/streaming", rb)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	msg := string(b)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, true, strings.Contains(msg, "'StreamingRequest.Signal' Error:Field validation for 'Signal' failed on the 'required' tag"))
}

func TestStreamingDeleteBadPayload(t *testing.T) {
	signal := ""
	recommendationItems := []models.ItemScore{}

	rb, err := createStreamingRequest("collaborative", signal, recommendationItems)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodDelete, "/v1/streaming", rb)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	msg := string(b)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, true, strings.Contains(msg, "'StreamingRequest.Signal' Error:Field validation for 'Signal' failed on the 'required' tag"))
}

func TestStreamingModelNotExist(t *testing.T) {
	signal := "100"
	recommendationItems := []models.ItemScore{
		{
			"item":  "111",
			"score": "0.6",
			"type":  "movie",
		},
		{
			"item":  "222",
			"score": "0.4",
			"type":  "movie",
		},
		{
			"item":  "555",
			"score": "0.16",
			"type":  "series",
		},
	}

	rb, err := createStreamingRequest("rintintin", signal, recommendationItems)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/v1/streaming", rb)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"message\":\"model with name rintintin not found\"}", string(b))
}

func TestStreamingUpdateModelNotExist(t *testing.T) {
	signal := "100"
	recommendationItems := []models.ItemScore{
		{
			"item":  "111",
			"score": "0.6",
			"type":  "movie",
		},
		{
			"item":  "222",
			"score": "0.4",
			"type":  "movie",
		},
		{
			"item":  "555",
			"score": "0.16",
			"type":  "series",
		},
	}

	rb, err := createStreamingRequest("titan", signal, recommendationItems)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPut, "/v1/streaming", rb)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"message\":\"model with name titan not found\"}", string(b))
}

func TestStreamingDeleteModelNotExist(t *testing.T) {
	signal := "100"
	recommendationItems := []models.ItemScore{
		{
			"item":  "111",
			"score": "0.6",
			"type":  "movie",
		},
		{
			"item":  "222",
			"score": "0.4",
			"type":  "movie",
		},
		{
			"item":  "555",
			"score": "0.16",
			"type":  "series",
		},
	}

	rb, err := createStreamingRequest("pine", signal, recommendationItems)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodDelete, "/v1/streaming", rb)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"message\":\"model pine not found\"}", string(b))
}

func TestStreamingUpdateData(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	if _, err := models.NewModel("collaborative", "", []string{"articleId"}, dbc); err != nil {
		t.FailNow()
	}

	signal := "543"
	recommendationItems := []models.ItemScore{
		{
			"item":  "111",
			"score": "0.6",
			"type":  "movie",
		},
		{
			"item":  "222",
			"score": "0.4",
			"type":  "movie",
		},
		{
			"item":  "555",
			"score": "0.16",
			"type":  "series",
		},
	}

	rb, err := createStreamingRequest("collaborative", signal, recommendationItems)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPut, "/v1/streaming", rb)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "{\"message\":\"signal 543 updated\"}", string(b))
}

func TestStreamingDeleteData(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	if _, err := models.NewModel("alcatraz", "", []string{"prisoner"}, dbc); err != nil {
		t.FailNow()
	}

	signal := "890"
	recommendationItems := []models.ItemScore{
		{
			"item":  "111",
			"score": "0.6",
			"type":  "movie",
		},
		{
			"item":  "222",
			"score": "0.4",
			"type":  "movie",
		},
		{
			"item":  "555",
			"score": "0.16",
			"type":  "series",
		},
	}

	rb, err := createStreamingRequest("alcatraz", signal, recommendationItems)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodDelete, "/v1/streaming", rb)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "{\"message\":\"signal 890 deleted\"}", string(b))
}

func TestBatchUploadDirect(t *testing.T) {

}

func TestBatchUploadDirectWithErrors(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	if _, err := models.NewModel("ham", "_", []string{"articleId","userId"}, dbc); err != nil {
		t.FailNow()
	}

	bd := make([]BatchData, 1)
	d := []models.ItemScore{
		{
			"item":  "111",
			"score": "0.6",
			"type":  "movie",
		},
		{
			"item":  "222",
			"score": "0.4",
			"type":  "movie",
		},
		{
			"item":  "555",
			"score": "0.16",
			"type":  "series",
		},
	}
	bd[0] = map[string][]models.ItemScore{
		"123": d,
		"124": d,
	}

	rb, err := createBatchRequestDirect("ham", bd)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/v1/batch", rb)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusCreated, code)
	assert.Equal(t, "{\"numberoflines\":\"2\",\"ErrorRecords\":{\"numberoflinesfailed\":\"2\",\"error\":[{\"1\":\"wrong format, the expected signal format must be articleId_userId\"},{\"2\":\"wrong format, the expected signal format must be articleId_userId\"}]}}", string(b))
}

func TestBatchUploadDirectNoErrors(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	if _, err := models.NewModel("pizza", "", []string{"articleId"}, dbc); err != nil {
		t.FailNow()
	}

	bd := make([]BatchData, 1)
	d := []models.ItemScore{
		{
			"item":  "111",
			"score": "0.6",
			"type":  "movie",
		},
		{
			"item":  "222",
			"score": "0.4",
			"type":  "movie",
		},
		{
			"item":  "555",
			"score": "0.16",
			"type":  "series",
		},
	}
	bd[0] = map[string][]models.ItemScore{
		"123": d,
		"124": d,
	}

	rb, err := createBatchRequestDirect("pizza", bd)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/v1/batch", rb)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusCreated, code)
	assert.Equal(t, "{\"numberoflines\":\"2\",\"ErrorRecords\":{\"numberoflinesfailed\":\"0\",\"error\":null}}", string(b))
}

func TestBatchUploadDirectModelNotExist(t *testing.T) {
	bd := make([]BatchData, 1)
	d := []models.ItemScore{
		{
			"item":  "111",
			"score": "0.6",
			"type":  "movie",
		},
		{
			"item":  "222",
			"score": "0.4",
			"type":  "movie",
		},
		{
			"item":  "555",
			"score": "0.16",
			"type":  "series",
		},
	}
	bd[0] = map[string][]models.ItemScore{
		"123": d,
	}

	rb, err := createBatchRequestDirect("karate", bd)
	if err != nil {
		t.Fail()
	}

	code, body, err := MockRequest(http.MethodPost, "/v1/batch", rb)
	if err != nil {
		t.Fail()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"message\":\"model with name karate not found\"}", string(b))
}

// CreateTestS3Bucket returns a bucket and defer a drop
func CreateTestS3Bucket(t *testing.T, bucket *db.S3Bucket, sess *session.Session) func() {
	s := db.NewS3Client(bucket, sess)
	s.CreateS3Bucket(&db.S3Bucket{Bucket: bucket.Bucket})
	return func() { s.DeleteS3Bucket(bucket) }
}

func
TestBatchUploadS3(t *testing.T) {
	var (
		s3TestEndpoint string = "localhost:4572"
		s3TestBucket   string = "test1"
		s3TestRegion   string = "eu-west-1"
		fileTest       string = "testdata/test_bulk_1key.jsonl"
		s3TestKey      string = "/" + fileTest
	)

	bucket := &db.S3Bucket{Bucket: s3TestBucket}
	sess := paws.NewAWSSession(s3TestRegion, s3TestEndpoint, true)
	s := db.NewS3Client(bucket, sess)

	dbc, c := GetTestRedisClient()
	defer c()

	if _, err := models.NewModel("batch", "", []string{"articleId"}, dbc); err != nil {
		t.FailNow()
	}

	drop := CreateTestS3Bucket(t, bucket, sess)
	defer drop()

	if err := s.UploadS3File(fileTest, bucket); err != nil {
		t.Fail()
	}

	rb, err := createBatchRequestLocation("batch", "s3://"+s3TestBucket+s3TestKey)
	if err != nil {
		t.Fail()
	}

	_, brsBody, err := MockRequest(http.MethodPost, "/v1/batch", rb)
	if err != nil {
		t.Fail()
	}

	var brs BatchBulkResponse
	if err := json.Unmarshal(brsBody.Bytes(), &brs); err != nil {
		t.Fail()
	}

	// wait for the upload to finish
	// due to the async goroutine, if we do not sleep the test
	// will loose the thread created by the endpoint
	time.Sleep(5 * time.Second)

	// do checks
	var srs BatchStatusResponse
	srsCode, srsBody, err := MockRequest(http.MethodGet, "/v1/batch/status/"+brs.BatchID, nil)
	if err != nil {
		t.Fail()
	}

	if err := json.Unmarshal(srsBody.Bytes(), &srs); err != nil {
		t.Fail()
	}

	switch srs.Status {
	case BulkSucceeded:
		assert.Equal(t, http.StatusOK, srsCode)
		assert.Equal(t, BulkSucceeded, srs.Status)
		return
	case BulkUploading:
	case BulkFailed:
	default:
		t.Fail()
	}
}

func TestBadBatchUploadS3(t *testing.T) {
	var (
		s3TestEndpoint string = "localhost:4572"
		s3TestBucket   string = "test1"
		s3TestRegion   string = "eu-west-1"
		fileTest       string = "testdata/test_bulk_1key.jsonl"
		s3TestKey      string = "/" + fileTest
	)

	bucket := &db.S3Bucket{Bucket: s3TestBucket}
	sess := paws.NewAWSSession(s3TestRegion, s3TestEndpoint, true)
	s := db.NewS3Client(bucket, sess)

	dbc, c := GetTestRedisClient()
	defer c()

	if _, err := models.NewModel("badbatch", "_", []string{"articleId","userId"}, dbc); err != nil {
		t.FailNow()
	}

	drop := CreateTestS3Bucket(t, bucket, sess)
	defer drop()

	if err := s.UploadS3File(fileTest, bucket); err != nil {
		t.Fail()
	}

	rb, err := createBatchRequestLocation("badbatch", "s3://"+s3TestBucket+s3TestKey)
	if err != nil {
		t.Fail()
	}

	_, brsBody, err := MockRequest(http.MethodPost, "/v1/batch", rb)
	if err != nil {
		t.Fail()
	}

	var brs BatchBulkResponse
	if err := json.Unmarshal(brsBody.Bytes(), &brs); err != nil {
		t.Fail()
	}

	// wait for the upload to finish
	// due to the async goroutine, if we do not sleep the test
	// will loose the thread created by the endpoint
	time.Sleep(5 * time.Second)

	// do checks
	var srs BatchStatusResponse
	srsCode, srsBody, err := MockRequest(http.MethodGet, "/v1/batch/status/"+brs.BatchID, nil)
	if err != nil {
		t.Fail()
	}

	if err := json.Unmarshal(srsBody.Bytes(), &srs); err != nil {
		t.Fail()
	}

	switch srs.Status {
	case BulkPartialUpload:
		assert.Equal(t, http.StatusOK, srsCode)
		assert.Equal(t, BulkPartialUpload, srs.Status)
		return
	case BulkSucceeded:
	case BulkUploading:
	case BulkFailed:
	default:
		t.Fail()
	}
}

func TestCorrectSignalFormat(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	if _, err := models.NewModel("signalFormat", "_", []string{"articleId","userId"}, dbc); err != nil {
		t.FailNow()
	}

	m, err := models.GetModel("signalFormat", dbc)
	if err != nil {
		t.FailNow()
	}

	tests := map[string]struct {
		input    string
		expected bool
	}{
		"correct": {
			input:    "11_22",
			expected: true,
		},
		"not correct 1": {
			input:    "11_33_33_33",
			expected: false,
		},
		"not correct 2": {
			input:    "11",
			expected: false,
		},
		"not correct 3": {
			input:    "11_",
			expected: false,
		},
		"not correct 4": {
			input:    "_11_",
			expected: false,
		},
		"not correct 5": {
			input:    "_11",
			expected: false,
		},
		"not correct 6": {
			input:    "_",
			expected: false,
		},
		"not correct 7": {
			input:    "11____",
			expected: false,
		},
		"not correct 8": {
			input:    "____11",
			expected: false,
		},
		"not correct 9": {
			input:    "",
			expected: false,
		},
	}
	for testName, test := range tests {
		t.Logf("Running test case %s", testName)
		o := m.CorrectSignalFormat(test.input)
		assert.Equal(t, test.expected, o)
	}
}
