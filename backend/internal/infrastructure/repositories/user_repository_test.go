package repositories_test

import (
	"context"
	"digital-signature-system/internal/domain/entities"
	"digital-signature-system/internal/infrastructure/repositories"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&entities.User{})
	require.NoError(t, err)

	return db
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := repositories.NewUserRepository(db)
	ctx := context.Background()

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

	err := repo.Create(ctx, user)
	assert.NoError(t, err)

	// Verify user was created
	savedUser, err := repo.GetByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, savedUser)
	assert.Equal(t, user.Username, savedUser.Username)
	assert.Equal(t, user.Email, savedUser.Email)
}

func TestUserRepository_GetByUsername(t *testing.T) {
	db := setupTestDB(t)
	repo := repositories.NewUserRepository(db)
	ctx := context.Background()

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

	err := repo.Create(ctx, user)
	assert.NoError(t, err)

	// Test get by username
	foundUser, err := repo.GetByUsername(ctx, "testuser")
	assert.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, user.ID, foundUser.ID)

	// Test non-existent username
	notFoundUser, err := repo.GetByUsername(ctx, "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, notFoundUser)
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := repositories.NewUserRepository(db)
	ctx := context.Background()

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

	err := repo.Create(ctx, user)
	assert.NoError(t, err)

	// Update user
	user.FullName = "Updated Name"
	err = repo.Update(ctx, user)
	assert.NoError(t, err)

	// Verify update
	updatedUser, err := repo.GetByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updatedUser.FullName)
}

func TestUserRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := repositories.NewUserRepository(db)
	ctx := context.Background()

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

	err := repo.Create(ctx, user)
	assert.NoError(t, err)

	// Delete user
	err = repo.Delete(ctx, user.ID)
	assert.NoError(t, err)

	// Verify deletion
	deletedUser, err := repo.GetByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.Nil(t, deletedUser)
}