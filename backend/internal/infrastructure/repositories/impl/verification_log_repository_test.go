package impl_test

import (
	"context"
	"digital-signature-system/internal/domain/entities"
	"digital-signature-system/internal/infrastructure/repositories/impl"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupVerificationLogTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&entities.User{}, &entities.Document{}, &entities.VerificationLog{})
	require.NoError(t, err)

	return db
}

func createTestDocument(t *testing.T, db *gorm.DB) *entities.Document {
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

	err = db.Create(doc).Error
	require.NoError(t, err)

	return doc
}

func TestVerificationLogRepository_Create(t *testing.T) {
	db := setupVerificationLogTestDB(t)
	repo := impl.NewVerificationLogRepository(db)
	ctx := context.Background()

	doc := createTestDocument(t, db)

	log := &entities.VerificationLog{
		ID:                 uuid.New().String(),
		DocumentID:         doc.ID,
		VerificationResult: "valid",
		VerifiedAt:         time.Now(),
		VerifierIP:         "127.0.0.1",
		Details:            `{"message": "Document is valid"}`,
	}

	err := repo.Create(ctx, log)
	assert.NoError(t, err)

	// Verify log was created
	savedLog, err := repo.GetByID(ctx, log.ID)
	assert.NoError(t, err)
	assert.NotNil(t, savedLog)
	assert.Equal(t, log.DocumentID, savedLog.DocumentID)
	assert.Equal(t, log.VerificationResult, savedLog.VerificationResult)
}

func TestVerificationLogRepository_GetByDocumentID(t *testing.T) {
	db := setupVerificationLogTestDB(t)
	repo := impl.NewVerificationLogRepository(db)
	ctx := context.Background()

	doc := createTestDocument(t, db)

	// Create multiple logs
	for i := 0; i < 3; i++ {
		log := &entities.VerificationLog{
			ID:                 uuid.New().String(),
			DocumentID:         doc.ID,
			VerificationResult: "valid",
			VerifiedAt:         time.Now().Add(time.Duration(-i) * time.Hour), // Different times
			VerifierIP:         "127.0.0." + string(rune(i+1)),
			Details:            `{"message": "Document is valid"}`,
		}
		err := repo.Create(ctx, log)
		require.NoError(t, err)
	}

	// Test getting logs by document ID
	logs, err := repo.GetByDocumentID(ctx, doc.ID)
	assert.NoError(t, err)
	assert.Len(t, logs, 3)

	// Verify logs are ordered by verified_at DESC
	assert.True(t, logs[0].VerifiedAt.After(logs[1].VerifiedAt))
	assert.True(t, logs[1].VerifiedAt.After(logs[2].VerifiedAt))
}