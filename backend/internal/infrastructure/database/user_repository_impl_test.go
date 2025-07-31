package database

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"digital-signature-system/internal/domain/entities"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Create table manually for SQLite compatibility
	err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			full_name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			role TEXT DEFAULT 'user',
			created_at DATETIME,
			updated_at DATETIME,
			is_active BOOLEAN DEFAULT true
		)
	`).Error
	if err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}

	// Create indexes
	db.Exec("CREATE INDEX idx_users_username ON users(username)")
	db.Exec("CREATE INDEX idx_users_email ON users(email)")

	return db
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &entities.User{
		ID:           uuid.New().String(),
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		FullName:     "Test User",
		Email:        "test@example.com",
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	err := repo.Create(ctx, user)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}

	// Verify user was created
	var count int64
	db.Model(&entities.User{}).Where("username = ?", "testuser").Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 user, got %d", count)
	}
}

func TestUserRepository_GetByUsername(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create test user
	user := &entities.User{
		ID:           uuid.New().String(),
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		FullName:     "Test User",
		Email:        "test@example.com",
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}
	db.Create(user)

	// Test GetByUsername
	result, err := repo.GetByUsername(ctx, "testuser")
	if err != nil {
		t.Errorf("GetByUsername() error = %v", err)
	}
	if result == nil {
		t.Error("Expected user, got nil")
	}
	if result.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got %s", result.Username)
	}

	// Test non-existent user
	result, err = repo.GetByUsername(ctx, "nonexistent")
	if err != nil {
		t.Errorf("GetByUsername() error = %v", err)
	}
	if result != nil {
		t.Error("Expected nil for non-existent user")
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create test user
	user := &entities.User{
		ID:           uuid.New().String(),
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		FullName:     "Test User",
		Email:        "test@example.com",
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}
	db.Create(user)

	// Test GetByEmail
	result, err := repo.GetByEmail(ctx, "test@example.com")
	if err != nil {
		t.Errorf("GetByEmail() error = %v", err)
	}
	if result == nil {
		t.Error("Expected user, got nil")
	}
	if result.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %s", result.Email)
	}
}