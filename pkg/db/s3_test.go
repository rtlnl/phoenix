package db

import (
	"bytes"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
)

const (
	s3TestEndpoint = "localhost:4572"
	s3TestBucket   = "test"
	s3TestRegion   = "eu-west-1"
)

func TestNewS3Client(t *testing.T) {
	s := NewS3Client(s3TestBucket, s3TestRegion, s3TestEndpoint, true)
	assert.NotNil(t, s)
}

func TestGetObject(t *testing.T) {
	s := NewS3Client(s3TestBucket, s3TestRegion, s3TestEndpoint, true)

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
	s := NewS3Client(s3TestBucket, s3TestRegion, s3TestEndpoint, true)

	f, err := s.GetObject("foo/bar2.txt")
	if err == nil {
		t.Failed()
	}

	assert.Equal(t, (*io.ReadCloser)(nil), f)
}

func TestExistsObject(t *testing.T) {
	s := NewS3Client(s3TestBucket, s3TestRegion, s3TestEndpoint, true)

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
	s := NewS3Client(s3TestBucket, s3TestRegion, s3TestEndpoint, true)

	// Key should not exists
	if s.ExistsObject("foo/bar2.txt") {
		t.Failed()
	}
}
