package db

import (
	"bytes"
	"io"
	"os"
	"testing"

	paws "github.com/rtlnl/data-personalization-api/pkg/aws"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
)

const (
	s3TestEndpoint = "localhost:4572"
	s3TestBucket   = "test"
	s3TestRegion   = "eu-west-1"
	s3TestKey      = "/foo/bar.txt"
)

func TestMain(m *testing.M) {
	tearUp()
	c := m.Run()
	tearDown()
	os.Exit(c)
}

func tearUp() {
	createS3Bucket()
}

func tearDown() {
	// Nothing here for now
}

func createS3Bucket() {
	sess := paws.NewAWSSession(s3TestRegion, s3TestEndpoint, true)
	sc := NewS3Client(s3TestBucket, sess)

	input := &s3.HeadBucketInput{
		Bucket: aws.String(s3TestBucket),
	}

	uploadFile := false
	if _, err := sc.Service.HeadBucket(input); err != nil {
		// create bucket if not
		sc.Service.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(s3TestBucket),
			ACL:    aws.String("public-read-write"),
		})
		uploadFile = true
	}

	// add files
	if uploadFile {
		input := &s3.PutObjectInput{
			Body:   bytes.NewReader([]byte("some-data")),
			Bucket: aws.String(s3TestBucket),
			Key:    aws.String(s3TestKey),
			ACL:    aws.String("public-read-write"),
		}

		if _, err := sc.Service.PutObject(input); err != nil {
			panic(err)
		}
	}
}

func TestNewS3Client(t *testing.T) {
	sess := paws.NewAWSSession(s3TestRegion, s3TestEndpoint, true)

	s := NewS3Client(s3TestBucket, sess)
	assert.NotNil(t, s)
}

func TestGetObject(t *testing.T) {
	sess := paws.NewAWSSession(s3TestRegion, s3TestEndpoint, true)
	s := NewS3Client(s3TestBucket, sess)

	_, err := s.Service.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s3TestBucket),
		Body:   bytes.NewReader([]byte("What is the meaning of life? 42.")),
		Key:    aws.String("foo/bar.txt"),
	})
	if err != nil {
		t.Failed()
	}

	f, err := s.GetObject("foo/bar.txt")
	if err != nil {
		t.Failed()
	}

	// convert body to string
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(*f)
	if err != nil {
		t.Failed()
	}

	assert.Equal(t, "What is the meaning of life? 42.", buf.String())
}

func TestGetObjectFails(t *testing.T) {
	sess := paws.NewAWSSession(s3TestRegion, s3TestEndpoint, true)
	s := NewS3Client(s3TestBucket, sess)

	f, err := s.GetObject("foo/bar2.txt")
	if err == nil {
		t.Failed()
	}

	assert.Equal(t, (*io.ReadCloser)(nil), f)
}

func TestExistsObject(t *testing.T) {
	sess := paws.NewAWSSession(s3TestRegion, s3TestEndpoint, true)
	s := NewS3Client(s3TestBucket, sess)

	_, err := s.Service.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s3TestBucket),
		Body:   bytes.NewReader([]byte("What is the meaning of life? 42.")),
		Key:    aws.String("foo/bar.txt"),
	})
	if err != nil {
		t.Failed()
	}

	if s.ExistsObject("foo/bar.txt") == false {
		t.Failed()
	}
}

func TestExistsObjectFails(t *testing.T) {
	sess := paws.NewAWSSession(s3TestRegion, s3TestEndpoint, true)
	s := NewS3Client(s3TestBucket, sess)

	// Key should not exists
	if s.ExistsObject("foo/bar2.txt") {
		t.Failed()
	}
}
