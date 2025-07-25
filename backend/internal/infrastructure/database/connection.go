package database

import (
	"fmt"
	"time"

	"digital-signature-system/internal/config"
	"digital-signature-system/internal/domain/entities"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewConnection(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set maximum number of open connections to the database
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	// Set maximum number of idle connections in the pool
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	// Set maximum amount of time a connection may be reused
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)

	// Set maximum amount of time a connection may be idle
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Minute)

	return db, nil
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&entities.User{},
		&entities.Session{},
		&entities.Document{},
		&entities.VerificationLog{},
	)
}