package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"digital-signature-system/internal/config"
	"digital-signature-system/internal/domain/entities"
)

func Initialize(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)

	// Configure GORM with PostgreSQL-friendly settings
	config := &gorm.Config{
		// Disable foreign key constraints during migration to avoid dependency issues
		DisableForeignKeyConstraintWhenMigrating: true,
		// Skip default transaction for better performance
		SkipDefaultTransaction: true,
		// Prepare statements for better performance
		PrepareStmt: true,
		// Use a more compatible logger level
		Logger: logger.Default.LogMode(logger.Silent),
	}
	
	db, err := gorm.Open(postgres.Open(dsn), config)
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

	// Use GORM migration with fallback approach
	if err := gormMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

func gormMigrate(db *gorm.DB) error {
	fmt.Println("Starting GORM migration...")
	
	// First, try the standard GORM AutoMigrate approach
	if err := tryStandardMigration(db); err != nil {
		fmt.Printf("Standard migration failed: %v\n", err)
		fmt.Println("Trying alternative migration approach...")
		return tryAlternativeMigration(db)
	}
	
	fmt.Println("GORM migration completed successfully!")
	return nil
}

func tryStandardMigration(db *gorm.DB) error {
	// Try standard AutoMigrate with all entities at once
	return db.AutoMigrate(
		&entities.User{},
		&entities.Session{},
		&entities.Document{},
		&entities.VerificationLog{},
	)
}

func tryAlternativeMigration(db *gorm.DB) error {
	// Alternative approach: migrate each entity individually with error handling
	entities := []interface{}{
		&entities.User{},
		&entities.Session{},
		&entities.Document{},
		&entities.VerificationLog{},
	}
	
	for _, entity := range entities {
		fmt.Printf("Migrating %T individually...\n", entity)
		if err := db.AutoMigrate(entity); err != nil {
			fmt.Printf("Individual migration failed for %T: %v\n", entity, err)
			// If individual migration also fails, try the migrator approach
			return tryMigratorApproach(db)
		}
	}
	
	return nil
}

func tryMigratorApproach(db *gorm.DB) error {
	fmt.Println("Using GORM Migrator interface...")
	
	migrator := db.Migrator()
	
	// Check if tables exist first
	entities := []interface{}{
		&entities.User{},
		&entities.Session{},
		&entities.Document{},
		&entities.VerificationLog{},
	}
	
	for _, entity := range entities {
		if !migrator.HasTable(entity) {
			fmt.Printf("Creating table for %T using Migrator...\n", entity)
			if err := migrator.CreateTable(entity); err != nil {
				return fmt.Errorf("failed to create table for %T: %w", entity, err)
			}
		} else {
			fmt.Printf("Table for %T already exists\n", entity)
		}
	}
	
	return nil
}

