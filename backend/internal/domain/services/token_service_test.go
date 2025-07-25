package services_test

import (
	"digital-signature-system/internal/domain/services"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTTokenService(t *testing.T) {
	// Create token service
	secretKey := "test-secret-key"
	issuer := "test-issuer"
	tokenService := services.NewJWTTokenService(secretKey, issuer)

	// Test GenerateToken
	userID := "user-123"
	username := "testuser"
	role := "user"
	duration := 1 * time.Hour

	token, err := tokenService.GenerateToken(userID, username, role, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Test ValidateToken
	claims, err := tokenService.ValidateToken(token)
	require.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, role, claims.Role)
	assert.Equal(t, issuer, claims.Issuer)
	assert.Equal(t, userID, claims.Subject)

	// Test with invalid token
	_, err = tokenService.ValidateToken("invalid-token")
	assert.Error(t, err)

	// Test with empty userID
	_, err = tokenService.GenerateToken("", username, role, duration)
	assert.Error(t, err)

	// Test with empty username
	_, err = tokenService.GenerateToken(userID, "", role, duration)
	assert.Error(t, err)

	// Test GenerateRefreshToken
	refreshToken, expiresAt, err := tokenService.GenerateRefreshToken()
	require.NoError(t, err)
	require.NotEmpty(t, refreshToken)
	assert.True(t, expiresAt.After(time.Now()))
	assert.True(t, expiresAt.Before(time.Now().Add(31*24*time.Hour))) // Should be around 30 days
}

func TestJWTTokenService_TokenExpiration(t *testing.T) {
	// Create token service
	secretKey := "test-secret-key"
	issuer := "test-issuer"
	tokenService := services.NewJWTTokenService(secretKey, issuer)

	// Generate token with very short expiration
	userID := "user-123"
	username := "testuser"
	role := "user"
	duration := 1 * time.Millisecond

	token, err := tokenService.GenerateToken(userID, username, role, duration)
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Validate token (should fail)
	_, err = tokenService.ValidateToken(token)
	assert.Error(t, err)
}