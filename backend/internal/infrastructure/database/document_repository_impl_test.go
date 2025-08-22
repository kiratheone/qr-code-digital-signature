package database

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"digital-signature-system/internal/domain/entities"
	"digital-signature-system/internal/domain/repositories"
)

// Helper function to create a pointer to a string
func stringPtr(s string) *string {
	return &s
}

const testUserID = "test-user-id"

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
			title TEXT,
			letter_number TEXT,
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
		UserID:        testUserID,
		Filename:      "test.pdf",
		Issuer:        "Test Issuer",
		Title:         stringPtr("Test Document Title"),
		LetterNumber:  stringPtr("LN-001"),
		DocumentHash:  "test-hash",
		SignatureData: "test-signature",
		QRCodeData:    "test-qr-data",
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
		Title:         stringPtr("Test Document Title 2"),
		LetterNumber:  stringPtr("LN-002"),
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
			Title:         stringPtr(fmt.Sprintf("Test Document %d", i+1)),
			LetterNumber:  stringPtr("LN-00X"),
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

func TestDocumentRepository_GetByID(t *testing.T) {
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
	docID := uuid.New().String()
	doc := &entities.Document{
		ID:            docID,
		UserID:        user.ID,
		Filename:      "test.pdf",
		Issuer:        "Test Issuer",
		Title:         stringPtr("Test Document Title 4"),
		LetterNumber:  stringPtr("LN-004"),
		DocumentHash:  "testhash123",
		SignatureData: "testsignature",
		QRCodeData:    "testqrcode",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		FileSize:      1024,
		Status:        "active",
	}
	db.Create(doc)

	// Test GetByID
	result, err := repo.GetByID(ctx, docID)
	if err != nil {
		t.Errorf("GetByID() error = %v", err)
	}
	if result == nil {
		t.Error("Expected document, got nil")
	}
	if result.ID != docID {
		t.Errorf("Expected ID %s, got %s", docID, result.ID)
	}
	if result.Filename != "test.pdf" {
		t.Errorf("Expected filename 'test.pdf', got %s", result.Filename)
	}

	// Test non-existent ID
	result, err = repo.GetByID(ctx, "nonexistent")
	if err != nil {
		t.Errorf("GetByID() error = %v", err)
	}
	if result != nil {
		t.Error("Expected nil for non-existent ID")
	}
}

func TestDocumentRepository_Update(t *testing.T) {
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
		Title:         stringPtr("Test Document Title 5"),
		LetterNumber:  stringPtr("LN-005"),
		DocumentHash:  "testhash123",
		SignatureData: "testsignature",
		QRCodeData:    "testqrcode",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		FileSize:      1024,
		Status:        "active",
	}
	db.Create(doc)

	// Update document
	doc.Status = "archived"
	doc.Issuer = "Updated Issuer"
	doc.UpdatedAt = time.Now()

	err := repo.Update(ctx, doc)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	// Verify update
	var updated entities.Document
	db.Where("id = ?", doc.ID).First(&updated)
	if updated.Status != "archived" {
		t.Errorf("Expected status 'archived', got %s", updated.Status)
	}
	if updated.Issuer != "Updated Issuer" {
		t.Errorf("Expected issuer 'Updated Issuer', got %s", updated.Issuer)
	}
}

func TestDocumentRepository_Delete(t *testing.T) {
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
	docID := uuid.New().String()
	doc := &entities.Document{
		ID:            docID,
		UserID:        user.ID,
		Filename:      "test.pdf",
		Issuer:        "Test Issuer",
		Title:         stringPtr("Test Document Title 6"),
		LetterNumber:  stringPtr("LN-006"),
		DocumentHash:  "testhash123",
		SignatureData: "testsignature",
		QRCodeData:    "testqrcode",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		FileSize:      1024,
		Status:        "active",
	}
	db.Create(doc)

	// Delete document
	err := repo.Delete(ctx, docID)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify deletion
	var count int64
	db.Model(&entities.Document{}).Where("id = ?", docID).Count(&count)
	if count != 0 {
		t.Errorf("Expected 0 documents after deletion, got %d", count)
	}

	// Test deleting non-existent document (should not error)
	err = repo.Delete(ctx, "nonexistent")
	if err != nil {
		t.Errorf("Delete() error for non-existent document = %v", err)
	}
}
