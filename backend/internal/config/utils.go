package config

import (
	"os"
	"strconv"
)

// getEnvInt gets an environment variable as an integer
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}