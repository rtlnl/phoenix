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

func TestGetDefault(t *testing.T) {
	tests := map[string]struct {
		input    string
		output   string
		expected string
	}{
		"get value": {
			input:    "hello",
			output:   "default",
			expected: "hello",
		},
		"get default": {
			input:    "",
			output:   "default",
			expected: "default",
		},
		"get default empty": {
			input:    "",
			output:   "",
			expected: "",
		},
	}
	for testName, test := range tests {
		t.Logf("Running test case %s", testName)
		o := GetDefault(test.input, test.output)

		assert.Equal(t, test.expected, o)
	}
}

func TestIsStringEmpty(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"not empty": {
			input:    "hello",
			expected: false,
		},
		"empty": {
			input:    "",
			expected: true,
		},
	}

	for testName, test := range tests {
		t.Logf("Running test case %s", testName)
		o := IsStringEmpty(test.input)

		assert.Equal(t, test.expected, o)
	}
}
