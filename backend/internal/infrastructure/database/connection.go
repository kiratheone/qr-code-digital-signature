package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"digital-signature-system/internal/config"
	"digital-signature-system/internal/domain/entities"
)

func Initialize(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set maximum number of open connections (as per design: max 10 connections)
	sqlDB.SetMaxOpenConns(10)
	// Set maximum number of idle connections
	sqlDB.SetMaxIdleConns(5)

	// Auto-migrate the schema
	if err := db.AutoMigrate(
		&entities.User{},
		&entities.Session{},
		&entities.Document{},
		&entities.VerificationLog{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}