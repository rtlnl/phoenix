package db

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/rs/zerolog/log"
)

// S3Client is the object that wraps around the official aws SDK
type S3Client struct {
	Service *s3.S3
	Bucket  string
}

type S3Bucket struct {
	Bucket string
	ACL    string
}

// NewS3Client is a wrapper object around the official aws s3 client
func NewS3Client(bucket *S3Bucket, sess *session.Session) *S3Client {
	return &S3Client{
		Service: s3.New(sess),
		Bucket:  bucket.Bucket,
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
		log.Warn().Msg(err.Error())
		return false
	}
	return true
}

// CreateS3Bucket creates a bucket
// It returns false if the bucket exists, true otherwise
func (c *S3Client) CreateS3Bucket(bucket *S3Bucket) (bool, error) {
	var err error

	// create bucket if doesn't exist
	if _, err = c.Service.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucket.Bucket),
	}); err != nil {

		_, err = c.Service.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(bucket.Bucket),
			ACL:    aws.String(bucket.ACL),
		})
		return true, err
	}

	return false, err
}

// DeleteS3Bucket creates a bucket
// It returns false if the bucket exists, true otherwise
func (c *S3Client) DeleteS3Bucket(bucket *S3Bucket) (bool, error) {
	var err error

	// delete bucket if exists
	if _, err = c.Service.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucket.Bucket),
	}); err == nil {

		// delete all objects
		err = c.DeleteS3AllObjects(bucket)
		if err != nil {
			return true, err
		}

		_, err = c.Service.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(bucket.Bucket),
		})
		return true, err
	}

	return false, err
}

// Delete all objects within a bucket (this is not the most efficient way)
func (c *S3Client) DeleteS3AllObjects(bucket *S3Bucket) error {

	// tetup BatchDeleteIterator to iterate through a list of objects.
	iter := s3manager.NewDeleteListIterator(c.Service, &s3.ListObjectsInput{
		Bucket: aws.String(bucket.Bucket),
	})

	for iter.Next() {
		o := iter.DeleteObject()
		_, err := c.Service.DeleteObject(&s3.DeleteObjectInput{Bucket: o.Object.Bucket, Key: o.Object.Key})

		if err != nil {
			return err
		}
	}

	return nil
}

// Delete all objects within a bucket (doesn't work)
func (c *S3Client) DeleteS3AllObjectsBatch(bucket *S3Bucket) error {

	// tetup BatchDeleteIterator to iterate through a list of objects.
	iter := s3manager.NewDeleteListIterator(c.Service, &s3.ListObjectsInput{
		Bucket: aws.String(bucket.Bucket),
	})

	// traverse iterator deleting each object
	return s3manager.NewBatchDeleteWithClient(c.Service).Delete(aws.BackgroundContext(), iter)
}
