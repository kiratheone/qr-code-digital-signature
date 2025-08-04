package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	Environment    string
	DBHost         string
	DBPort         string
	DBName         string
	DBUser         string
	DBPassword     string
	DBSSLMode      string
	JWTSecret      string
	PrivateKey     string
	PublicKey      string
	PrivateKeyPath string
	PublicKeyPath  string
}

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	config := &Config{
		Port:           getEnv("PORT", "8000"),
		Environment:    getEnv("ENVIRONMENT", "development"),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBName:         getEnv("DB_NAME", "digital_signature"),
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPassword:     getEnv("DB_PASSWORD", "password"),
		DBSSLMode:      getEnv("DB_SSL_MODE", "disable"),
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key"),
		PrivateKey:     getEnv("PRIVATE_KEY", ""),
		PublicKey:      getEnv("PUBLIC_KEY", ""),
		PrivateKeyPath: getEnv("PRIVATE_KEY_PATH", "private_key.pem"),
		PublicKeyPath:  getEnv("PUBLIC_KEY_PATH", "public_key.pem"),
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}