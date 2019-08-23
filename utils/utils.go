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
