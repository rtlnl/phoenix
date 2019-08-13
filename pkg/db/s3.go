package db

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3Client is the object that wraps around the official aws SDK
type S3Client struct {
	Service *s3.S3
	Bucket  string
}

// NewS3Client is a wrapper object around the official aws s3 client
func NewS3Client(bucket string, sess *session.Session) *S3Client {
	return &S3Client{
		Service: s3.New(sess),
		Bucket:  bucket,
	}
}

// GetObject returns the specified object-key from the selected bucket
// TODO: We need to be smarted here! Currently we download the whole file.
// Better if we manage to separate it in different chunks
func (c *S3Client) GetObject(key string) (*io.ReadCloser, error) {
	o, err := c.Service.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(c.Bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, err
	}

	return &o.Body, nil
}

// ExistsObject test if the specified key exists in the bucket
// It returns true if the key exists, false otherwise
func (c *S3Client) ExistsObject(key string) bool {
	// There is no official method to test if a key exists or not.
	// To avoid downloading the object, we require the metadata
	// information instead
	req, _ := c.Service.HeadObjectRequest(&s3.HeadObjectInput{
		Bucket: aws.String(c.Bucket),
		Key:    aws.String(key),
	})

	if err := req.Send(); err != nil {
		return false
	}
	return true
}
