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
func NewS3Client(bucket, region string) *S3Client {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:           aws.String(region),
		S3ForcePathStyle: aws.Bool(true),
		DisableSSL:       aws.Bool(true),               // TODO: Remove this
		Endpoint:         aws.String("localhost:4572"), // TODO: Remove this
	}))

	return &S3Client{
		Service: s3.New(sess),
		Bucket:  bucket,
	}
}

// GetObject returns the specified object-key from the selected bucket
// TODO: improve the function to stream it in chunks to optimize the reading process
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
