package config

import (
	"os"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Security SecurityConfig
}

type ServerConfig struct {
	Port string
	Host string
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
			Port: getEnv("PORT", "8000"),
			Host: getEnv("HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "password"),
			DBName:          getEnv("DB_NAME", "digital_signature"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 20),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvInt("DB_CONN_MAX_LIFETIME", 60), // 60 minutes
			ConnMaxIdleTime: getEnvInt("DB_CONN_MAX_IDLE_TIME", 30), // 30 minutes
		},
		Security: SecurityConfig{
			PrivateKey:      getEnv("PRIVATE_KEY", ""),
			PublicKey:       getEnv("PUBLIC_KEY", ""),
			JWTSecret:       getEnv("JWT_SECRET", "your-secret-key"),
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