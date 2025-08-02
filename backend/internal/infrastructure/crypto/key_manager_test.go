package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewKeyManager(t *testing.T) {
	// Save original environment
	originalPrivateKey := os.Getenv("RSA_PRIVATE_KEY")
	originalPublicKey := os.Getenv("RSA_PUBLIC_KEY")
	originalPrivateKeyPath := os.Getenv("PRIVATE_KEY_PATH")
	originalPublicKeyPath := os.Getenv("PUBLIC_KEY_PATH")

	defer func() {
		// Restore original environment
		os.Setenv("RSA_PRIVATE_KEY", originalPrivateKey)
		os.Setenv("RSA_PUBLIC_KEY", originalPublicKey)
		os.Setenv("PRIVATE_KEY_PATH", originalPrivateKeyPath)
		os.Setenv("PUBLIC_KEY_PATH", originalPublicKeyPath)
	}()

	tests := []struct {
		name           string
		setupEnv       func()
		wantErr        bool
		errContains    string
		checkKeyManager func(t *testing.T, km *KeyManager)
	}{
		{
			name: "load from environment variables",
			setupEnv: func() {
				privateKeyPEM, publicKeyPEM := createTestKeyPairPEM(t)
				os.Setenv("RSA_PRIVATE_KEY", privateKeyPEM)
				os.Setenv("RSA_PUBLIC_KEY", publicKeyPEM)
				os.Unsetenv("PRIVATE_KEY_PATH")
				os.Unsetenv("PUBLIC_KEY_PATH")
			},
			wantErr: false,
			checkKeyManager: func(t *testing.T, km *KeyManager) {
				assert.NotNil(t, km.GetPrivateKey())
				assert.NotNil(t, km.GetPublicKey())
				assert.NotEmpty(t, km.GetKeyID())
				assert.NoError(t, km.ValidateKeys())
			},
		},
		{
			name: "load from files with custom paths",
			setupEnv: func() {
				os.Unsetenv("RSA_PRIVATE_KEY")
				os.Unsetenv("RSA_PUBLIC_KEY")
				os.Setenv("PRIVATE_KEY_PATH", "../../../../private_key.pem")
				os.Setenv("PUBLIC_KEY_PATH", "../../../../public_key.pem")
			},
			wantErr: false,
			checkKeyManager: func(t *testing.T, km *KeyManager) {
				assert.NotNil(t, km.GetPrivateKey())
				assert.NotNil(t, km.GetPublicKey())
				assert.NotEmpty(t, km.GetKeyID())
				assert.NoError(t, km.ValidateKeys())
			},
		},
		{
			name: "load from default file paths",
			setupEnv: func() {
				os.Unsetenv("RSA_PRIVATE_KEY")
				os.Unsetenv("RSA_PUBLIC_KEY")
				os.Unsetenv("PRIVATE_KEY_PATH")
				os.Unsetenv("PUBLIC_KEY_PATH")
				// This will fail because default paths don't exist in test environment
			},
			wantErr:     true,
			errContains: "failed to load private key from file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()

			km, err := NewKeyManager()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, km)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, km)
				if tt.checkKeyManager != nil {
					tt.checkKeyManager(t, km)
				}
			}
		})
	}
}

func TestNewKeyManagerFromEnv(t *testing.T) {
	privateKeyPEM, publicKeyPEM := createTestKeyPairPEM(t)

	tests := []struct {
		name         string
		privateKeyEnv string
		publicKeyEnv  string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "valid PEM keys",
			privateKeyEnv: privateKeyPEM,
			publicKeyEnv:  publicKeyPEM,
			wantErr:      false,
		},
		{
			name:         "valid base64 encoded keys",
			privateKeyEnv: base64.StdEncoding.EncodeToString([]byte(privateKeyPEM)),
			publicKeyEnv:  base64.StdEncoding.EncodeToString([]byte(publicKeyPEM)),
			wantErr:      false,
		},
		{
			name:         "private key only (derive public key)",
			privateKeyEnv: privateKeyPEM,
			publicKeyEnv:  "",
			wantErr:      false,
		},
		{
			name:         "empty private key",
			privateKeyEnv: "",
			publicKeyEnv:  publicKeyPEM,
			wantErr:      true,
			errContains:  "private key environment variable is empty",
		},
		{
			name:         "invalid private key",
			privateKeyEnv: "invalid_key_data",
			publicKeyEnv:  publicKeyPEM,
			wantErr:      true,
			errContains:  "failed to decode base64 private key",
		},
		{
			name:         "invalid base64 private key",
			privateKeyEnv: "invalid_base64!@#",
			publicKeyEnv:  publicKeyPEM,
			wantErr:      true,
			errContains:  "failed to decode base64 private key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			km, err := NewKeyManagerFromEnv(tt.privateKeyEnv, tt.publicKeyEnv)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, km)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, km)
				assert.NotNil(t, km.GetPrivateKey())
				assert.NotNil(t, km.GetPublicKey())
				assert.NotEmpty(t, km.GetKeyID())
				assert.NoError(t, km.ValidateKeys())
			}
		})
	}
}

func TestNewKeyManagerFromFiles(t *testing.T) {
	tests := []struct {
		name           string
		privateKeyPath string
		publicKeyPath  string
		wantErr        bool
		errContains    string
	}{
		{
			name:           "valid key files",
			privateKeyPath: "../../../../private_key.pem",
			publicKeyPath:  "../../../../public_key.pem",
			wantErr:        false,
		},
		{
			name:           "nonexistent private key file",
			privateKeyPath: "nonexistent_private.pem",
			publicKeyPath:  "../../../../public_key.pem",
			wantErr:        true,
			errContains:    "failed to load private key from file",
		},
		{
			name:           "nonexistent public key file",
			privateKeyPath: "../../../../private_key.pem",
			publicKeyPath:  "nonexistent_public.pem",
			wantErr:        true,
			errContains:    "failed to load public key from file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			km, err := NewKeyManagerFromFiles(tt.privateKeyPath, tt.publicKeyPath)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, km)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, km)
				assert.NotNil(t, km.GetPrivateKey())
				assert.NotNil(t, km.GetPublicKey())
				assert.NotEmpty(t, km.GetKeyID())
				assert.NoError(t, km.ValidateKeys())
			}
		})
	}
}

func TestKeyManager_ValidateKeys(t *testing.T) {
	// Create a valid key manager
	validKM := createTestKeyManager(t)

	// Create invalid key managers for testing
	privateKey1, _ := rsa.GenerateKey(rand.Reader, 2048)
	privateKey2, _ := rsa.GenerateKey(rand.Reader, 2048)

	tests := []struct {
		name        string
		keyManager  *KeyManager
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid key pair",
			keyManager: validKM,
			wantErr:    false,
		},
		{
			name: "nil private key",
			keyManager: &KeyManager{
				privateKey: nil,
				publicKey:  &privateKey1.PublicKey,
			},
			wantErr:     true,
			errContains: "private key is nil",
		},
		{
			name: "nil public key",
			keyManager: &KeyManager{
				privateKey: privateKey1,
				publicKey:  nil,
			},
			wantErr:     true,
			errContains: "public key is nil",
		},
		{
			name: "mismatched key pair",
			keyManager: &KeyManager{
				privateKey: privateKey1,
				publicKey:  &privateKey2.PublicKey,
			},
			wantErr:     true,
			errContains: "public key does not match private key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.keyManager.ValidateKeys()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestKeyManager_GenerateNewKeyPair(t *testing.T) {
	km := createTestKeyManager(t)

	tests := []struct {
		name        string
		keySize     int
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid 2048-bit key",
			keySize: 2048,
			wantErr: false,
		},
		{
			name:    "valid 4096-bit key",
			keySize: 4096,
			wantErr: false,
		},
		{
			name:        "invalid key size (too small)",
			keySize:     1024,
			wantErr:     true,
			errContains: "key size must be at least 2048 bits",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyPair, err := km.GenerateNewKeyPair(tt.keySize)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, keyPair)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, keyPair)
				assert.NotEmpty(t, keyPair.PrivateKey)
				assert.NotEmpty(t, keyPair.PublicKey)
				assert.NotEmpty(t, keyPair.KeyID)
				assert.Equal(t, "RSA-2048", keyPair.Algorithm)
				assert.True(t, strings.Contains(keyPair.PrivateKey, "-----BEGIN PRIVATE KEY-----"))
				assert.True(t, strings.Contains(keyPair.PublicKey, "-----BEGIN PUBLIC KEY-----"))

				// Verify the generated key pair is valid
				tempKM, err := NewKeyManagerFromEnv(keyPair.PrivateKey, keyPair.PublicKey)
				assert.NoError(t, err)
				assert.NoError(t, tempKM.ValidateKeys())
			}
		})
	}
}

func TestKeyManager_RotateKeys(t *testing.T) {
	km := createTestKeyManager(t)

	// Generate a new key pair for rotation
	newKeyPair, err := km.GenerateNewKeyPair(2048)
	require.NoError(t, err)

	tests := []struct {
		name        string
		newKeyPair  *KeyPair
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid key rotation",
			newKeyPair: newKeyPair,
			wantErr:    false,
		},
		{
			name:        "nil key pair",
			newKeyPair:  nil,
			wantErr:     true,
			errContains: "new key pair cannot be nil",
		},
		{
			name: "invalid private key in new pair",
			newKeyPair: &KeyPair{
				PrivateKey: "invalid_private_key",
				PublicKey:  newKeyPair.PublicKey,
				KeyID:      "test_key",
				CreatedAt:  time.Now(),
			},
			wantErr:     true,
			errContains: "failed to parse new private key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh key manager for each test
			testKM := createTestKeyManager(t)
			
			err := testKM.RotateKeys(tt.newKeyPair)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				// Key should remain unchanged on error - only check if newKeyPair is not nil
				if tt.newKeyPair != nil {
					assert.NotEqual(t, tt.newKeyPair.KeyID, testKM.GetKeyID())
				}
			} else {
				assert.NoError(t, err)
				// Key should be updated
				assert.Equal(t, tt.newKeyPair.KeyID, testKM.GetKeyID())
				assert.NoError(t, testKM.ValidateKeys())
				// The key ID should be different from original (unless by coincidence)
				// We'll just verify the rotation worked by checking the key ID matches the new pair
			}
		})
	}
}

func TestKeyManager_ExportKeyPairForStorage(t *testing.T) {
	km := createTestKeyManager(t)

	keyPair, err := km.ExportKeyPairForStorage()

	assert.NoError(t, err)
	assert.NotNil(t, keyPair)
	assert.NotEmpty(t, keyPair.PrivateKey)
	assert.NotEmpty(t, keyPair.PublicKey)
	assert.Equal(t, km.GetKeyID(), keyPair.KeyID)
	assert.Equal(t, "RSA-2048", keyPair.Algorithm)
	assert.True(t, strings.Contains(keyPair.PrivateKey, "-----BEGIN PRIVATE KEY-----"))
	assert.True(t, strings.Contains(keyPair.PublicKey, "-----BEGIN PUBLIC KEY-----"))

	// Verify exported key pair can be used to create a new key manager
	newKM, err := NewKeyManagerFromEnv(keyPair.PrivateKey, keyPair.PublicKey)
	assert.NoError(t, err)
	assert.NoError(t, newKM.ValidateKeys())
	assert.Equal(t, km.GetKeyID(), newKM.GetKeyID())
}

func TestKeyManager_GetMethods(t *testing.T) {
	km := createTestKeyManager(t)

	// Test all getter methods
	assert.NotNil(t, km.GetPrivateKey())
	assert.NotNil(t, km.GetPublicKey())
	assert.NotEmpty(t, km.GetKeyID())
	assert.False(t, km.GetCreatedAt().IsZero())

	// Test new security methods
	assert.Equal(t, 2048, km.GetKeySize())
	assert.True(t, km.GetKeyAge() >= 0)
	assert.False(t, km.ShouldRotateKey(time.Hour)) // Should not need rotation within an hour

	// Verify the keys are valid RSA keys
	assert.IsType(t, &rsa.PrivateKey{}, km.GetPrivateKey())
	assert.IsType(t, &rsa.PublicKey{}, km.GetPublicKey())
}

func TestKeyManager_SecurityFeatures(t *testing.T) {
	km := createTestKeyManager(t)

	tests := []struct {
		name        string
		maxAge      time.Duration
		shouldRotate bool
	}{
		{
			name:         "should not rotate within 1 hour",
			maxAge:       time.Hour,
			shouldRotate: false,
		},
		{
			name:         "should rotate after 1 nanosecond",
			maxAge:       time.Nanosecond,
			shouldRotate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Wait a tiny bit to ensure some time passes
			time.Sleep(time.Microsecond)
			result := km.ShouldRotateKey(tt.maxAge)
			assert.Equal(t, tt.shouldRotate, result)
		})
	}
}

func TestKeyManager_ClearKeys(t *testing.T) {
	km := createTestKeyManager(t)

	// Verify keys exist before clearing
	assert.NotNil(t, km.GetPrivateKey())
	assert.NotNil(t, km.GetPublicKey())
	assert.NotEmpty(t, km.GetKeyID())

	// Clear keys
	km.ClearKeys()

	// Verify keys are cleared
	assert.Nil(t, km.GetPrivateKey())
	assert.Nil(t, km.GetPublicKey())
	assert.Empty(t, km.GetKeyID())
	assert.Equal(t, 0, km.GetKeySize())
}

func TestKeyManager_ValidateKeysWithKeySize(t *testing.T) {
	// Create a key manager with a small key for testing
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024) // Below minimum
	require.NoError(t, err)

	km := &KeyManager{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
		keyID:      "test_key",
		createdAt:  time.Now(),
	}

	err = km.ValidateKeys()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key size 1024 bits is below minimum security requirement")
}

func TestParsePrivateKeyFromPEM(t *testing.T) {
	// Create test keys
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Test PKCS#8 format
	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	pkcs8PEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	})

	// Test PKCS#1 format
	pkcs1PEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	tests := []struct {
		name        string
		pemData     string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid PKCS#8 format",
			pemData: string(pkcs8PEM),
			wantErr: false,
		},
		{
			name:    "valid PKCS#1 format",
			pemData: string(pkcs1PEM),
			wantErr: false,
		},
		{
			name:        "invalid PEM format",
			pemData:     "not a pem file",
			wantErr:     true,
			errContains: "failed to decode PEM block",
		},
		{
			name: "invalid key data",
			pemData: `-----BEGIN PRIVATE KEY-----
aW52YWxpZF9rZXlfZGF0YQ==
-----END PRIVATE KEY-----`,
			wantErr:     true,
			errContains: "failed to parse private key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := parsePrivateKeyFromPEM(tt.pemData)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, key)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, key)
				assert.IsType(t, &rsa.PrivateKey{}, key)
			}
		})
	}
}

func TestParsePublicKeyFromPEM(t *testing.T) {
	// Create test key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)
	validPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	tests := []struct {
		name        string
		pemData     string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid public key",
			pemData: string(validPEM),
			wantErr: false,
		},
		{
			name:        "invalid PEM format",
			pemData:     "not a pem file",
			wantErr:     true,
			errContains: "failed to decode PEM block",
		},
		{
			name: "invalid key data",
			pemData: `-----BEGIN PUBLIC KEY-----
aW52YWxpZF9rZXlfZGF0YQ==
-----END PUBLIC KEY-----`,
			wantErr:     true,
			errContains: "failed to parse public key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := parsePublicKeyFromPEM(tt.pemData)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, key)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, key)
				assert.IsType(t, &rsa.PublicKey{}, key)
			}
		})
	}
}

func TestKeyRotationIntegration(t *testing.T) {
	// Create initial key manager
	km := createTestKeyManager(t)
	originalKeyID := km.GetKeyID()

	// Generate new key pair (this will create a different key)
	newKeyPair, err := km.GenerateNewKeyPair(2048)
	require.NoError(t, err)

	// Rotate to new keys
	err = km.RotateKeys(newKeyPair)
	require.NoError(t, err)

	// Verify rotation was successful
	assert.Equal(t, newKeyPair.KeyID, km.GetKeyID())
	// Since we generated a new key pair, the key ID should be different
	// (unless by extreme coincidence, which is virtually impossible with RSA keys)
	assert.NotEqual(t, originalKeyID, km.GetKeyID())
	assert.NoError(t, km.ValidateKeys())

	// Export the rotated keys
	exportedKeyPair, err := km.ExportKeyPairForStorage()
	require.NoError(t, err)
	assert.Equal(t, newKeyPair.KeyID, exportedKeyPair.KeyID)

	// Create new key manager from exported keys
	newKM, err := NewKeyManagerFromEnv(exportedKeyPair.PrivateKey, exportedKeyPair.PublicKey)
	require.NoError(t, err)
	assert.Equal(t, km.GetKeyID(), newKM.GetKeyID())
	assert.NoError(t, newKM.ValidateKeys())
}

// Helper functions

func createTestKeyManager(t *testing.T) *KeyManager {
	km, err := NewKeyManagerFromFiles("../../../../private_key.pem", "../../../../public_key.pem")
	require.NoError(t, err)
	return km
}

func createTestKeyPairPEM(t *testing.T) (string, string) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Private key to PEM
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Public key to PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(privateKeyPEM), string(publicKeyPEM)
}

func createTestKeyFiles(t *testing.T) (string, string) {
	tempDir := t.TempDir()

	privateKeyPEM, publicKeyPEM := createTestKeyPairPEM(t)

	privateKeyPath := filepath.Join(tempDir, "private_key.pem")
	publicKeyPath := filepath.Join(tempDir, "public_key.pem")

	err := os.WriteFile(privateKeyPath, []byte(privateKeyPEM), 0600)
	require.NoError(t, err)

	err = os.WriteFile(publicKeyPath, []byte(publicKeyPEM), 0644)
	require.NoError(t, err)

	return privateKeyPath, publicKeyPath
}