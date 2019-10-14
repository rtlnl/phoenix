package internal

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	paws "github.com/rtlnl/data-personalization-api/pkg/aws"

	"github.com/rtlnl/data-personalization-api/models"
	"github.com/rtlnl/data-personalization-api/pkg/db"
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
	// get aerospike client
	ac, c := GetTestAerospikeClient()
	defer c()

	truncate := CreateTestModel(t, ac, "rtl_nieuws", "fancy", "", []string{"articleId"}, false)
	defer truncate()

	signal := "100"
	recommendationItems := []string{"1", "2", "3", "4"}

	rb, err := createStreamingRequest("rtl_nieuws", "fancy", signal, recommendationItems)
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

func TestStreamingBadSignal(t *testing.T) {
	// get aerospike client
	ac, c := GetTestAerospikeClient()
	defer c()

	truncate := CreateTestModel(t, ac, "rtl_nieuws", "fancy", "_", []string{"articleId", "userId"}, false)
	defer truncate()

	signal := "100"
	recommendationItems := []string{"1", "2", "3", "4"}

	rb, err := createStreamingRequest("rtl_nieuws", "fancy", signal, recommendationItems)
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
	assert.Equal(t, "{\"message\":\"the expected signal format must be articleId_userId\"}", string(b))
}

func TestStreamingBadPayload(t *testing.T) {
	signal := ""
	recommendationItems := []string{}

	rb, err := createStreamingRequest("", "", signal, recommendationItems)
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

	msg := string(b)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, true, strings.Contains(msg, "'StreamingRequest.Signal' Error:Field validation for 'Signal' failed on the 'required' tag"))
}

func TestStreamingUpdateBadPayload(t *testing.T) {
	signal := ""
	recommendationItems := []string{}

	rb, err := createStreamingRequest("", "", signal, recommendationItems)
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

	msg := string(b)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, true, strings.Contains(msg, "'StreamingRequest.Signal' Error:Field validation for 'Signal' failed on the 'required' tag"))
}

func TestStreamingDeleteBadPayload(t *testing.T) {
	signal := ""
	recommendationItems := []string{}

	rb, err := createStreamingRequest("", "", signal, recommendationItems)
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

	msg := string(b)

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, true, strings.Contains(msg, "'StreamingRequest.Signal' Error:Field validation for 'Signal' failed on the 'required' tag"))
}

func TestStreamingPublishedModel(t *testing.T) {
	// get aerospike client
	ac, c := GetTestAerospikeClient()
	defer c()

	truncate := CreateTestModel(t, ac, "rtl_nieuws", "hello", "", []string{"articleId"}, true)
	defer truncate()

	signal := "100"
	recommendationItems := []string{"1", "2", "3", "4"}

	rb, err := createStreamingRequest("rtl_nieuws", "hello", signal, recommendationItems)
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
	assert.Equal(t, "{\"message\":\"you cannot add data on already published models. stage it first\"}", string(b))
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

	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"message\":\"key pizza does not exist\"}", string(b))
}

func TestStreamingUpdateModelNotExist(t *testing.T) {
	signal := "100"
	recommendationItems := []string{"1", "2", "3", "4"}

	rb, err := createStreamingRequest("pasta", "pizza", signal, recommendationItems)
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

	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"message\":\"key pizza does not exist\"}", string(b))
}

func TestStreamingDeleteModelNotExist(t *testing.T) {
	signal := "100"
	recommendationItems := []string{"1", "2", "3", "4"}

	rb, err := createStreamingRequest("pasta", "pizza", signal, recommendationItems)
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

	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"message\":\"key pizza does not exist\"}", string(b))
}

func TestStreamingUpdateData(t *testing.T) {
	// get aerospike client
	ac, c := GetTestAerospikeClient()
	defer c()

	truncate := CreateTestModel(t, ac, "rtl_nieuws", "fancy", "", []string{"articleId"}, false)
	defer truncate()

	signal := "543"
	recommendationItems := []string{"6", "7", "8", "9"}

	rb, err := createStreamingRequest("rtl_nieuws", "fancy", signal, recommendationItems)
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
	assert.Equal(t, "{\"message\":\"signal 543 updated\"}", string(b))
}

func TestStreamingUpdateDataPublishedModel(t *testing.T) {
	// get aerospike client
	ac, c := GetTestAerospikeClient()
	defer c()

	truncate := CreateTestModel(t, ac, "rtl_nieuws", "homepage", "", []string{"articleId"}, true)
	defer truncate()

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

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "{\"message\":\"you cannot update data on already published models. stage it first\"}", string(b))
}

func TestStreamingDeleteData(t *testing.T) {
	// get aerospike client
	ac, c := GetTestAerospikeClient()
	defer c()

	truncate := CreateTestModel(t, ac, "rtl_nieuws", "burger", "", []string{"articleId"}, false)
	defer truncate()

	signal := "890"
	recommendationItems := []string{"6", "7", "8", "9"}

	rb, err := createStreamingRequest("rtl_nieuws", "burger", signal, recommendationItems)
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
	assert.Equal(t, "{\"message\":\"signal 890 deleted\"}", string(b))
}

func TestStreamingDeleteDataPublishedModel(t *testing.T) {
	// get aerospike client
	ac, c := GetTestAerospikeClient()
	defer c()

	truncate := CreateTestModel(t, ac, "rtl_nieuws", "banana", "", []string{"articleId"}, true)
	defer truncate()

	signal := "100"
	recommendationItems := []string{"6", "7", "8", "9"}

	rb, err := createStreamingRequest("rtl_nieuws", "banana", signal, recommendationItems)
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

	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "{\"message\":\"you cannot delete data on already published models. stage it first\"}", string(b))
}

func TestBatchUploadDirect(t *testing.T) {

}

func TestBatchUploadDirectWithErrors(t *testing.T) {
	// get aerospike client
	ac, c := GetTestAerospikeClient()
	defer c()

	truncate := CreateTestModel(t, ac, "rtl_nieuws", "bread", "_", []string{"articleId", "userId"}, false)
	defer truncate()

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

	rb, err := createBatchRequestDirect("rtl_nieuws", "bread", bd)
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

	assert.Equal(t, http.StatusCreated, code)
	assert.Equal(t, "{\"numberoflines\":\"2\",\"ErrorRecords\":{\"numberoflinesfailed\":\"2\",\"error\":[{\"lines\":\"1 ,2\",\"reason\":\"wrong format, the expected signal format must be articleId_userId\"}]}}", string(b))
}

func TestBatchUploadDirectNoErrors(t *testing.T) {
	// get aerospike client
	ac, c := GetTestAerospikeClient()
	defer c()

	truncate := CreateTestModel(t, ac, "rtl_nieuws", "bread", "", []string{"articleId"}, false)
	defer truncate()

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

	rb, err := createBatchRequestDirect("rtl_nieuws", "bread", bd)
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

	assert.Equal(t, http.StatusCreated, code)
	assert.Equal(t, "{\"numberoflines\":\"2\",\"ErrorRecords\":{\"numberoflinesfailed\":\"0\",\"error\":null}}", string(b))
}

func TestBatchUploadDirectModelPublished(t *testing.T) {
	// get aerospike client
	ac, c := GetTestAerospikeClient()
	defer c()

	truncate := CreateTestModel(t, ac, "rtl_nieuws", "bread", "", []string{"articleId"}, true)
	defer truncate()

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

	rb, err := createBatchRequestDirect("rtl_nieuws", "bread", bd)
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

	rb, err := createBatchRequestDirect("pasta", "pizza", bd)
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

	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "{\"message\":\"model with publicationPoint pasta and campaign pizza not found\"}", string(b))
}

func TestStripS3URL(t *testing.T) {
	l := "s3://test-bucket/foo/bar/hello.csv"
	expectedBucket := "test-bucket"
	expectedKey := "foo/bar/hello.csv"

	bucket, key := StripS3URL(l)

	assert.Equal(t, expectedBucket, bucket)
	assert.Equal(t, expectedKey, key)
}

// CreateTestS3Bucket returns a bucket and defer a drop
func CreateTestS3Bucket(t *testing.T, bucket *db.S3Bucket, sess *session.Session) func() {
	s := db.NewS3Client(bucket, sess)
	s.CreateS3Bucket(&db.S3Bucket{Bucket: bucket.Bucket})
	return func() { s.DeleteS3Bucket(bucket) }
}

func TestBatchUploadS3(t *testing.T) {
	var (
		s3TestEndpoint string = "localhost:4572"
		s3TestBucket   string = "test1"
		s3TestRegion   string = "eu-west-1"
		fileTest       string = "testdata/test_bulk_1key.jsonl"
		s3TestKey      string = "/" + fileTest
	)

	var (
		srsCode int = 0
		n       int = 5
		sleep   int = 2
	)

	var srsBody *bytes.Buffer
	var srs BatchStatusResponse

	bucket := &db.S3Bucket{Bucket: s3TestBucket}
	sess := paws.NewAWSSession(s3TestRegion, s3TestEndpoint, true)
	s := db.NewS3Client(bucket, sess)

	// get aerospike client
	ac, c := GetTestAerospikeClient()
	defer c()

	truncate := CreateTestModel(t, ac, "rtl_nieuws", "bread", "", []string{"articleId"}, false)
	defer truncate()

	drop := CreateTestS3Bucket(t, bucket, sess)
	defer drop()

	if err := s.UploadS3File(fileTest, bucket); err != nil {
		t.Fail()
	}

	rb, err := createBatchRequestLocation("rtl_nieuws", "bread", "s3://"+s3TestBucket+s3TestKey)
	if err != nil {
		t.Fail()
	}

	_, brsBody, err := MockRequest(http.MethodPost, "/batch", rb)
	if err != nil {
		t.Fail()
	}

	var brs BatchBulkResponse
	if err := json.Unmarshal(brsBody.Bytes(), &brs); err != nil {
		t.Fail()
	}

	// do checks
	for i := 0; i < n; i++ {
		srsCode, srsBody, err = MockRequest(http.MethodGet, "/batch/status/"+brs.BatchID, nil)
		if err != nil {
			t.Fail()
		}

		if err := json.Unmarshal(srsBody.Bytes(), &srs); err != nil {
			t.Fail()
		}

		// cannot use case because the break
		if srs.Status == "SUCCEEDED" {
			break
		} else if srs.Status == "FAILED" {
			t.Fail()
		} else {

			// UPLOADING: wait sleep seconds and failt if it took more than 5 iterations
			time.Sleep(time.Duration(sleep) * time.Second)
			if i == n {
				t.Fail()
			}
		}
	}

	assert.Equal(t, http.StatusOK, srsCode)
	assert.Equal(t, "SUCCEEDED", srs.Status)
}

func TestBadBatchUploadS3(t *testing.T) {
	var (
		s3TestEndpoint string = "localhost:4572"
		s3TestBucket   string = "test1"
		s3TestRegion   string = "eu-west-1"
		fileTest       string = "testdata/test_bulk_1key.jsonl"
		s3TestKey      string = "/" + fileTest
	)

	var (
		srsCode int = 0
		n       int = 5
		sleep   int = 2
	)

	var srsBody *bytes.Buffer
	var srs BatchStatusResponse

	bucket := &db.S3Bucket{Bucket: s3TestBucket}
	sess := paws.NewAWSSession(s3TestRegion, s3TestEndpoint, true)
	s := db.NewS3Client(bucket, sess)

	// get aerospike client
	ac, c := GetTestAerospikeClient()
	defer c()

	truncate := CreateTestModel(t, ac, "rtl_nieuws", "bread", "_", []string{"articleId", "userId"}, false)
	defer truncate()

	drop := CreateTestS3Bucket(t, bucket, sess)
	defer drop()

	if err := s.UploadS3File(fileTest, bucket); err != nil {
		t.Fail()
	}

	rb, err := createBatchRequestLocation("rtl_nieuws", "bread", "s3://"+s3TestBucket+s3TestKey)
	if err != nil {
		t.Fail()
	}

	_, brsBody, err := MockRequest(http.MethodPost, "/batch", rb)
	if err != nil {
		t.Fail()
	}

	var brs BatchBulkResponse
	if err := json.Unmarshal(brsBody.Bytes(), &brs); err != nil {
		t.Fail()
	}

	// do checks
	for i := 0; i < n; i++ {
		srsCode, srsBody, err = MockRequest(http.MethodGet, "/batch/status/"+brs.BatchID, nil)
		if err != nil {
			t.Fail()
		}

		if err := json.Unmarshal(srsBody.Bytes(), &srs); err != nil {
			t.Fail()
		}

		// cannot use case because the break
		if srs.Status == "SUCCEEDED" {
			break
		} else if srs.Status == "FAILED" {
			t.Fail()
		} else {

			// UPLOADING: wait sleep seconds and failt if it took more than 5 iterations
			time.Sleep(time.Duration(sleep) * time.Second)
			if i == n {
				t.Fail()
			}
		}
	}

	assert.Equal(t, http.StatusOK, srsCode)
	assert.Equal(t, "SUCCEEDED", srs.Status)
}
