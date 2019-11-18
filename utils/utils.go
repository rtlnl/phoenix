package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// GetEnv will set an env variable with a default if the variable is not
// found in the system. Used for testing purposes
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// GetDefault returns def if the parameter val is empty
func GetDefault(val, def string) string {
	if val != "" {
		return val
	}
	return def
}

// StringInSlice checks if a string is found in a slice
func StringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

// StripS3URL returns the bucket and the key from a s3 url location
func StripS3URL(URL string) (string, string) {
	bucketTmp := strings.Replace(URL, "s3://", "", -1)
	bucket := bucketTmp[:strings.IndexByte(bucketTmp, '/')]
	key := strings.TrimPrefix(URL, fmt.Sprintf("s3://%s/", bucket))
	return bucket, key
}

// RemoveEmptyValueInSlice returns a slice without empty strings
func RemoveEmptyValueInSlice(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

// IsStringEmpty checks if the input is empty
func IsStringEmpty(m string) bool {
	if m == "" {
		return true
	}
	return false
}

// SerializeObject returns the JSON string representation of the object
func SerializeObject(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// RemoveElemFromSlice removes an element from the slice. Not super efficient but it does the job
func RemoveElemFromSlice(v string, l []string) []string {
	var res []string
	for _, elem := range l {
		if elem == v {
			continue
		}
		res = append(res, elem)
	}
	return res
}