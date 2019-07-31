package db

import (
	"bytes"
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
		t.FailNow()
	}

	f, err := s.GetObject("foo/bar.txt")
	if err != nil {
		t.FailNow()
	}

	// convert body to string
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(*f)
	if err != nil {
		t.FailNow()
	}

	assert.Equal(t, "What is the meaning of life? 42.", buf.String())
}
