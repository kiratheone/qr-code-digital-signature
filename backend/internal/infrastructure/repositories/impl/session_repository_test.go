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

func setupSessionTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&entities.User{}, &entities.Session{})
	require.NoError(t, err)

	return db
}

func TestSessionRepository_Create(t *testing.T) {
	db := setupSessionTestDB(t)
	repo := impl.NewSessionRepository(db)
	ctx := context.Background()

	user := createTestUser(t, db)

	session := &entities.Session{
		ID:           uuid.New().String(),
		UserID:       user.ID,
		SessionToken: "session-token-123",
		RefreshToken: "refresh-token-123",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
	}

	err := repo.Create(ctx, session)
	assert.NoError(t, err)

	// Verify session was created
	savedSession, err := repo.GetByToken(ctx, session.SessionToken)
	assert.NoError(t, err)
	assert.NotNil(t, savedSession)
	assert.Equal(t, session.UserID, savedSession.UserID)
	assert.Equal(t, session.RefreshToken, savedSession.RefreshToken)
}

func TestSessionRepository_Delete(t *testing.T) {
	db := setupSessionTestDB(t)
	repo := impl.NewSessionRepository(db)
	ctx := context.Background()

	user := createTestUser(t, db)

	session := &entities.Session{
		ID:           uuid.New().String(),
		UserID:       user.ID,
		SessionToken: "session-token-123",
		RefreshToken: "refresh-token-123",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
	}

	err := repo.Create(ctx, session)
	assert.NoError(t, err)

	// Delete session
	err = repo.Delete(ctx, session.SessionToken)
	assert.NoError(t, err)

	// Verify deletion
	deletedSession, err := repo.GetByToken(ctx, session.SessionToken)
	assert.NoError(t, err)
	assert.Nil(t, deletedSession)
}

func TestSessionRepository_DeleteByUserID(t *testing.T) {
	db := setupSessionTestDB(t)
	repo := impl.NewSessionRepository(db)
	ctx := context.Background()

	user := createTestUser(t, db)

	// Create multiple sessions for the same user
	for i := 0; i < 3; i++ {
		session := &entities.Session{
			ID:           uuid.New().String(),
			UserID:       user.ID,
			SessionToken: "session-token-" + uuid.New().String()[:8],
			RefreshToken: "refresh-token-" + uuid.New().String()[:8],
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now(),
			LastAccessed: time.Now(),
		}
		err := repo.Create(ctx, session)
		require.NoError(t, err)
	}

	// Delete all sessions for the user
	err := repo.DeleteByUserID(ctx, user.ID)
	assert.NoError(t, err)

	// Verify all sessions were deleted
	var count int64
	err = db.Model(&entities.Session{}).Where("user_id = ?", user.ID).Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}