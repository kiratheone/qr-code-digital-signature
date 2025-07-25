package impl_test

import (
	"context"
	"digital-signature-system/internal/domain/entities"
	"digital-signature-system/internal/domain/repositories"
	"digital-signature-system/internal/infrastructure/repositories/impl"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDocumentTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&entities.User{}, &entities.Document{})
	require.NoError(t, err)

	return db
}

func createTestUser(t *testing.T, db *gorm.DB) *entities.User {
	user := &entities.User{
		ID:           uuid.New().String(),
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		FullName:     "Test User",
		Email:        "test@example.com",
		Role:         "user",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := db.Create(user).Error
	require.NoError(t, err)

	return user
}

func TestDocumentRepository_Create(t *testing.T) {
	db := setupDocumentTestDB(t)
	repo := impl.NewDocumentRepository(db)
	ctx := context.Background()

	user := createTestUser(t, db)

	doc := &entities.Document{
		ID:            uuid.New().String(),
		UserID:        user.ID,
		Filename:      "test.pdf",
		Issuer:        "Test Issuer",
		DocumentHash:  "hash123",
		SignatureData: "signature123",
		QRCodeData:    "qrdata123",
		FileSize:      1024,
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := repo.Create(ctx, doc)
	assert.NoError(t, err)

	// Verify document was created
	savedDoc, err := repo.GetByID(ctx, doc.ID)
	assert.NoError(t, err)
	assert.NotNil(t, savedDoc)
	assert.Equal(t, doc.Filename, savedDoc.Filename)
	assert.Equal(t, doc.DocumentHash, savedDoc.DocumentHash)
}

func TestDocumentRepository_GetByUserID(t *testing.T) {
	db := setupDocumentTestDB(t)
	docRepo := impl.NewDocumentRepository(db)
	ctx := context.Background()

	user := createTestUser(t, db)

	// Create multiple documents
	for i := 0; i < 5; i++ {
		doc := &entities.Document{
			ID:            uuid.New().String(),
			UserID:        user.ID,
			Filename:      "test" + uuid.New().String()[:8] + ".pdf",
			Issuer:        "Test Issuer",
			DocumentHash:  "hash" + uuid.New().String()[:8],
			SignatureData: "signature" + uuid.New().String()[:8],
			QRCodeData:    "qrdata" + uuid.New().String()[:8],
			FileSize:      1024,
			Status:        "active",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		err := docRepo.Create(ctx, doc)
		require.NoError(t, err)
	}

	// Test getting documents with pagination
	filter := repositories.DocumentFilter{
		Limit:  3,
		Offset: 0,
	}

	docs, total, err := docRepo.GetByUserID(ctx, user.ID, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, docs, 3)

	// Test with offset
	filter.Offset = 3
	docs, total, err = docRepo.GetByUserID(ctx, user.ID, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, docs, 2)
}

func TestDocumentRepository_Update(t *testing.T) {
	db := setupDocumentTestDB(t)
	docRepo := impl.NewDocumentRepository(db)
	ctx := context.Background()

	user := createTestUser(t, db)

	doc := &entities.Document{
		ID:            uuid.New().String(),
		UserID:        user.ID,
		Filename:      "test.pdf",
		Issuer:        "Test Issuer",
		DocumentHash:  "hash123",
		SignatureData: "signature123",
		QRCodeData:    "qrdata123",
		FileSize:      1024,
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := docRepo.Create(ctx, doc)
	assert.NoError(t, err)

	// Update document
	doc.Issuer = "Updated Issuer"
	err = docRepo.Update(ctx, doc)
	assert.NoError(t, err)

	// Verify update
	updatedDoc, err := docRepo.GetByID(ctx, doc.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Issuer", updatedDoc.Issuer)
}

func TestDocumentRepository_Delete(t *testing.T) {
	db := setupDocumentTestDB(t)
	docRepo := impl.NewDocumentRepository(db)
	ctx := context.Background()

	user := createTestUser(t, db)

	doc := &entities.Document{
		ID:            uuid.New().String(),
		UserID:        user.ID,
		Filename:      "test.pdf",
		Issuer:        "Test Issuer",
		DocumentHash:  "hash123",
		SignatureData: "signature123",
		QRCodeData:    "qrdata123",
		FileSize:      1024,
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := docRepo.Create(ctx, doc)
	assert.NoError(t, err)

	// Delete document
	err = docRepo.Delete(ctx, doc.ID)
	assert.NoError(t, err)

	// Verify deletion
	deletedDoc, err := docRepo.GetByID(ctx, doc.ID)
	assert.NoError(t, err)
	assert.Nil(t, deletedDoc)
}