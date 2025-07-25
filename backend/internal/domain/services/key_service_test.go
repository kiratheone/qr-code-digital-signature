package services_test

import (
	"digital-signature-system/internal/domain/services"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileKeyService(t *testing.T) {
	// Create temporary directory for test keys
	tempDir, err := os.MkdirTemp("", "key-service-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	privateKeyPath := filepath.Join(tempDir, "private.pem")
	publicKeyPath := filepath.Join(tempDir, "public.pem")

	// Create key service
	keyService := services.NewFileKeyService(privateKeyPath, publicKeyPath)

	// Test GenerateAndSaveKeys
	privateKey, publicKey, err := keyService.GenerateAndSaveKeys(2048)
	require.NoError(t, err)
	require.NotEmpty(t, privateKey)
	require.NotEmpty(t, publicKey)

	// Verify files were created
	assert.FileExists(t, privateKeyPath)
	assert.FileExists(t, publicKeyPath)

	// Test LoadKeys
	loadedPrivateKey, loadedPublicKey, err := keyService.LoadKeys()
	require.NoError(t, err)
	assert.Equal(t, privateKey, loadedPrivateKey)
	assert.Equal(t, publicKey, loadedPublicKey)

	// Test RotateKeys
	rotatedPrivateKey, rotatedPublicKey, err := keyService.RotateKeys(2048)
	require.NoError(t, err)
	require.NotEmpty(t, rotatedPrivateKey)
	require.NotEmpty(t, rotatedPublicKey)

	// Keys should be different after rotation
	assert.NotEqual(t, privateKey, rotatedPrivateKey)
	assert.NotEqual(t, publicKey, rotatedPublicKey)

	// Verify backup files were created
	files, err := os.ReadDir(tempDir)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(files), 4) // 2 current files + at least 2 backups
}

func TestEnvironmentKeyService(t *testing.T) {
	// Define environment variable names for test
	privateKeyEnv := "TEST_PRIVATE_KEY"
	publicKeyEnv := "TEST_PUBLIC_KEY"

	// Clear environment variables before test
	os.Unsetenv(privateKeyEnv)
	os.Unsetenv(publicKeyEnv)

	// Create key service
	keyService := services.NewEnvironmentKeyService(privateKeyEnv, publicKeyEnv)

	// Test GenerateAndSaveKeys
	privateKey, publicKey, err := keyService.GenerateAndSaveKeys(2048)
	require.NoError(t, err)
	require.NotEmpty(t, privateKey)
	require.NotEmpty(t, publicKey)

	// Verify environment variables were set
	assert.Equal(t, privateKey, os.Getenv(privateKeyEnv))
	assert.Equal(t, publicKey, os.Getenv(publicKeyEnv))

	// Test LoadKeys
	loadedPrivateKey, loadedPublicKey, err := keyService.LoadKeys()
	require.NoError(t, err)
	assert.Equal(t, privateKey, loadedPrivateKey)
	assert.Equal(t, publicKey, loadedPublicKey)

	// Test RotateKeys
	rotatedPrivateKey, rotatedPublicKey, err := keyService.RotateKeys(2048)
	require.NoError(t, err)
	require.NotEmpty(t, rotatedPrivateKey)
	require.NotEmpty(t, rotatedPublicKey)

	// Keys should be different after rotation
	assert.NotEqual(t, privateKey, rotatedPrivateKey)
	assert.NotEqual(t, publicKey, rotatedPublicKey)

	// Verify environment variables were updated
	assert.Equal(t, rotatedPrivateKey, os.Getenv(privateKeyEnv))
	assert.Equal(t, rotatedPublicKey, os.Getenv(publicKeyEnv))

	// Clean up
	os.Unsetenv(privateKeyEnv)
	os.Unsetenv(publicKeyEnv)
}

func TestValidateRSAKeys(t *testing.T) {
	// Generate a valid key pair
	signatureService := &services.RSASignatureService{}
	privateKeyPEM, publicKeyPEM, err := signatureService.GenerateKeyPair(2048)
	require.NoError(t, err)

	// Test ValidateRSAPrivateKey with valid key
	err = services.ValidateRSAPrivateKey(privateKeyPEM)
	assert.NoError(t, err)

	// Test ValidateRSAPublicKey with valid key
	err = services.ValidateRSAPublicKey(publicKeyPEM)
	assert.NoError(t, err)

	// Test ValidateRSAPrivateKey with invalid key
	err = services.ValidateRSAPrivateKey("invalid key")
	assert.Error(t, err)

	// Test ValidateRSAPublicKey with invalid key
	err = services.ValidateRSAPublicKey("invalid key")
	assert.Error(t, err)
}