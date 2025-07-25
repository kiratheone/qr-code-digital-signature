package services_test

import (
	"digital-signature-system/internal/domain/services"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestBCryptPasswordService(t *testing.T) {
	// Create password service with minimum cost for faster tests
	passwordService := services.NewBCryptPasswordService(bcrypt.MinCost)

	// Test HashPassword
	password := "test-password"
	hash, err := passwordService.HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	// Test VerifyPassword with correct password
	valid, err := passwordService.VerifyPassword(password, hash)
	require.NoError(t, err)
	assert.True(t, valid)

	// Test VerifyPassword with incorrect password
	valid, err = passwordService.VerifyPassword("wrong-password", hash)
	require.NoError(t, err)
	assert.False(t, valid)

	// Test with empty password
	_, err = passwordService.HashPassword("")
	assert.Error(t, err)

	// Test verify with empty password
	_, err = passwordService.VerifyPassword("", hash)
	assert.Error(t, err)

	// Test verify with empty hash
	_, err = passwordService.VerifyPassword(password, "")
	assert.Error(t, err)
}

func TestBCryptPasswordService_DefaultCost(t *testing.T) {
	// Test with invalid cost (too low)
	passwordService := services.NewBCryptPasswordService(-1)
	
	// Should still work with default cost
	password := "test-password"
	hash, err := passwordService.HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hash)
	
	// Test with invalid cost (too high)
	passwordService = services.NewBCryptPasswordService(100)
	
	// Should still work with default cost
	hash, err = passwordService.HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hash)
}