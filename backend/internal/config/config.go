package config

import (
	"os"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Security SecurityConfig
}

type ServerConfig struct {
	Port        string
	Host        string
	Environment string
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int // in minutes
	ConnMaxIdleTime int // in minutes
}

type SecurityConfig struct {
	PrivateKey      string
	PublicKey       string
	JWTSecret       string
	KeyRotationDays int
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:        getEnv("PORT", "8000"),
			Host:        getEnv("HOST", "0.0.0.0"),
			Environment: getEnv("ENVIRONMENT", "development"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getSecret("db_password", "DB_PASSWORD", "password"),
			DBName:          getEnv("DB_NAME", "digital_signature"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 20),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvInt("DB_CONN_MAX_LIFETIME", 60), // 60 minutes
			ConnMaxIdleTime: getEnvInt("DB_CONN_MAX_IDLE_TIME", 30), // 30 minutes
		},
		Security: SecurityConfig{
			PrivateKey:      getSecret("private_key", "PRIVATE_KEY", ""),
			PublicKey:       getSecret("public_key", "PUBLIC_KEY", ""),
			JWTSecret:       getSecret("jwt_secret", "JWT_SECRET", "your-secret-key"),
			KeyRotationDays: getEnvInt("KEY_ROTATION_DAYS", 90),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}



// GetString returns a string configuration value
func (c *Config) GetString(key, defaultValue string) string {
	return getEnv(key, defaultValue)
}

// GetInt returns an integer configuration value
func (c *Config) GetInt(key string, defaultValue int) int {
	return getEnvInt(key, defaultValue)
}

// GetInt64 returns an int64 configuration value
func (c *Config) GetInt64(key string, defaultValue int64) int64 {
	return getEnvInt64(key, defaultValue)
}

// GetFloat64 returns a float64 configuration value
func (c *Config) GetFloat64(key string, defaultValue float64) float64 {
	return getEnvFloat64(key, defaultValue)
}

// GetDuration returns a duration configuration value
func (c *Config) GetDuration(key string, defaultValue time.Duration) time.Duration {
	return getEnvDuration(key, defaultValue)
}