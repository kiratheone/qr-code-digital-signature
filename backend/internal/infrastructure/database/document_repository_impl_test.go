package database

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"digital-signature-system/internal/domain/entities"
	"digital-signature-system/internal/domain/repositories"
)

func setupDocumentTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Create tables manually for SQLite compatibility
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

	err = db.Exec(`
		CREATE TABLE documents (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			filename TEXT NOT NULL,
			issuer TEXT NOT NULL,
			document_hash TEXT NOT NULL,
			signature_data TEXT NOT NULL,
			qr_code_data TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			file_size INTEGER,
			status TEXT DEFAULT 'active',
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`).Error
	if err != nil {
		t.Fatalf("failed to create documents table: %v", err)
	}

	// Create indexes
	db.Exec("CREATE INDEX idx_documents_user_id ON documents(user_id)")
	db.Exec("CREATE INDEX idx_documents_hash ON documents(document_hash)")
	db.Exec("CREATE INDEX idx_documents_created_at ON documents(created_at)")

	return db
}

func TestDocumentRepository_Create(t *testing.T) {
	db := setupDocumentTestDB(t)
	repo := NewDocumentRepository(db)
	ctx := context.Background()

	// Create test user first
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

	doc := &entities.Document{
		ID:            uuid.New().String(),
		UserID:        user.ID,
		Filename:      "test.pdf",
		Issuer:        "Test Issuer",
		DocumentHash:  "testhash123",
		SignatureData: "testsignature",
		QRCodeData:    "testqrcode",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		FileSize:      1024,
		Status:        "active",
	}

	err := repo.Create(ctx, doc)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}

	// Verify document was created
	var count int64
	db.Model(&entities.Document{}).Where("filename = ?", "test.pdf").Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 document, got %d", count)
	}
}

func TestDocumentRepository_GetByHash(t *testing.T) {
	db := setupDocumentTestDB(t)
	repo := NewDocumentRepository(db)
	ctx := context.Background()

	// Create test user first
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

	// Create test document
	doc := &entities.Document{
		ID:            uuid.New().String(),
		UserID:        user.ID,
		Filename:      "test.pdf",
		Issuer:        "Test Issuer",
		DocumentHash:  "testhash123",
		SignatureData: "testsignature",
		QRCodeData:    "testqrcode",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		FileSize:      1024,
		Status:        "active",
	}
	db.Create(doc)

	// Test GetByHash
	result, err := repo.GetByHash(ctx, "testhash123")
	if err != nil {
		t.Errorf("GetByHash() error = %v", err)
	}
	if result == nil {
		t.Error("Expected document, got nil")
	}
	if result.DocumentHash != "testhash123" {
		t.Errorf("Expected hash 'testhash123', got %s", result.DocumentHash)
	}

	// Test non-existent hash
	result, err = repo.GetByHash(ctx, "nonexistent")
	if err != nil {
		t.Errorf("GetByHash() error = %v", err)
	}
	if result != nil {
		t.Error("Expected nil for non-existent hash")
	}
}

func TestDocumentRepository_GetByUserID(t *testing.T) {
	db := setupDocumentTestDB(t)
	repo := NewDocumentRepository(db)
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

	// Create test documents
	for i := 0; i < 3; i++ {
		doc := &entities.Document{
			ID:            uuid.New().String(),
			UserID:        user.ID,
			Filename:      "test.pdf",
			Issuer:        "Test Issuer",
			DocumentHash:  "testhash" + string(rune(i)),
			SignatureData: "testsignature",
			QRCodeData:    "testqrcode",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			FileSize:      1024,
			Status:        "active",
		}
		db.Create(doc)
	}

	// Test GetByUserID with pagination
	filter := repositories.DocumentFilter{
		Page:     1,
		PageSize: 2,
		Status:   "active",
	}

	docs, total, err := repo.GetByUserID(ctx, user.ID, filter)
	if err != nil {
		t.Errorf("GetByUserID() error = %v", err)
	}
	if len(docs) != 2 {
		t.Errorf("Expected 2 documents, got %d", len(docs))
	}
	if total != 3 {
		t.Errorf("Expected total 3, got %d", total)
	}
}