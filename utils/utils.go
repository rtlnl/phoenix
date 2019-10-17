package utils

import "os"

// GetEnv will set an env variable with a default if the variable is not
// found in the system. Used for testing purposes
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Checks if a string is found in a slice
func StringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

// The objects coming from Aerospike have type []interface{}. This function converts
// the Bins in the appropriate type for consistency
func ConvertInterfaceToList(bins interface{}) []string {
	var list []string
	newBins := bins.([]interface{})
	for _, bin := range newBins {
		list = append(list, bin.(string))
	}
	return list
}
