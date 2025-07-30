package config

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
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

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvFloat64(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// getSecret reads a Docker secret from file or falls back to environment variable
func getSecret(secretName, envKey, defaultValue string) string {
	// First try to read from Docker secret file
	secretPath := "/run/secrets/" + secretName
	if data, err := ioutil.ReadFile(secretPath); err == nil {
		return strings.TrimSpace(string(data))
	}
	
	// Fall back to environment variable
	return getEnv(envKey, defaultValue)
}