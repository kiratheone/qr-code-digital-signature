package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	Environment    string
	BaseURL        string
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
	CORSOrigins    string
}

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	config := &Config{
		Port:           getEnv("PORT", "8000"),
		Environment:    getEnv("ENVIRONMENT", "development"),
		BaseURL:        getEnv("APP_BASE_URL", "http://localhost:3000"),
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
		CORSOrigins:    getEnv("CORS_ORIGINS", "https://sign.arikachmad.com,https://sign-api.arikachmad.com,http://localhost:3000,http://localhost:8065"),
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetCORSOrigins returns CORS origins as a slice
func (c *Config) GetCORSOrigins() []string {
	if c.CORSOrigins == "" {
		return []string{}
	}
	origins := strings.Split(c.CORSOrigins, ",")
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}
	return origins
}
