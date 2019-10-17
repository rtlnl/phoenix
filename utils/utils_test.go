package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripS3URL(t *testing.T) {
	l := "s3://test-bucket/foo/bar/hello.csv"
	expectedBucket := "test-bucket"
	expectedKey := "foo/bar/hello.csv"

	bucket, key := StripS3URL(l)

	assert.Equal(t, expectedBucket, bucket)
	assert.Equal(t, expectedKey, key)
}

func TestStringInSlice(t *testing.T) {
	s := "hello"
	l := []string{"hello", "world"}

	assert.Equal(t, true, StringInSlice(s, l))
	assert.Equal(t, false, StringInSlice("banana", l))
}

func TestConvertInterfaceToList(t *testing.T) {
	l := []interface{}{"hello", "world"}
	ls := []string{"hello", "world"}

	assert.ElementsMatch(t, ls, ConvertInterfaceToList(l))
}
