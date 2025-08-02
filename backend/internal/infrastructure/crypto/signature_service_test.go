package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSignatureService(t *testing.T) {
	tests := []struct {
		name           string
		privateKeyPath string
		publicKeyPath  string
		wantErr        bool
		errContains    string
	}{
		{
			name:           "valid keys",
			privateKeyPath: "../../../../private_key.pem",
			publicKeyPath:  "../../../../public_key.pem",
			wantErr:        false,
		},
		{
			name:           "invalid private key path",
			privateKeyPath: "nonexistent.pem",
			publicKeyPath:  "../../../../public_key.pem",
			wantErr:        true,
			errContains:    "failed to load private key",
		},
		{
			name:           "invalid public key path",
			privateKeyPath: "../../../../private_key.pem",
			publicKeyPath:  "nonexistent.pem",
			wantErr:        true,
			errContains:    "failed to load public key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewSignatureService(tt.privateKeyPath, tt.publicKeyPath)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, service)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
				assert.NotNil(t, service.privateKey)
				assert.NotNil(t, service.publicKey)
			}
		})
	}
}

func TestSignatureService_CalculateDocumentHash(t *testing.T) {
	service := createTestSignatureService(t)

	tests := []struct {
		name         string
		documentData []byte
		expected     []byte
	}{
		{
			name:         "simple text",
			documentData: []byte("Hello, World!"),
			expected:     func() []byte { h := sha256.Sum256([]byte("Hello, World!")); return h[:] }(),
		},
		{
			name:         "empty data",
			documentData: []byte(""),
			expected:     func() []byte { h := sha256.Sum256([]byte("")); return h[:] }(),
		},
		{
			name:         "binary data",
			documentData: []byte{0x00, 0x01, 0x02, 0xFF},
			expected:     func() []byte { h := sha256.Sum256([]byte{0x00, 0x01, 0x02, 0xFF}); return h[:] }(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CalculateDocumentHash(tt.documentData)
			assert.Equal(t, tt.expected, result)
			assert.Len(t, result, 32) // SHA-256 produces 32-byte hash
		})
	}
}

func TestSignatureService_SignDocument(t *testing.T) {
	service := createTestSignatureService(t)

	// Create a proper SHA-256 hash for testing
	testData := []byte("test document content")
	validHash := service.CalculateDocumentHash(testData)

	tests := []struct {
		name         string
		documentHash []byte
		wantErr      bool
		errContains  string
	}{
		{
			name:         "valid hash",
			documentHash: validHash,
			wantErr:      false,
		},
		{
			name:         "empty hash",
			documentHash: []byte(""),
			wantErr:      true,
			errContains:  "document hash cannot be empty",
		},
		{
			name:         "nil hash",
			documentHash: nil,
			wantErr:      true,
			errContains:  "document hash cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signatureData, err := service.SignDocument(tt.documentHash)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, signatureData)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, signatureData)
				assert.NotEmpty(t, signatureData.Signature)
				assert.Equal(t, tt.documentHash, signatureData.Hash)
				assert.Equal(t, "RSA-PSS-SHA256", signatureData.Algorithm)
			}
		})
	}
}

func TestSignatureService_VerifySignature(t *testing.T) {
	service := createTestSignatureService(t)
	
	// Create a proper SHA-256 hash for testing
	testData := []byte("test document content")
	testHash := service.CalculateDocumentHash(testData)

	// Create a valid signature for testing
	validSignature, err := service.SignDocument(testHash)
	require.NoError(t, err)

	tests := []struct {
		name          string
		documentHash  []byte
		signatureData *SignatureData
		wantErr       bool
		errContains   string
	}{
		{
			name:          "valid signature",
			documentHash:  testHash,
			signatureData: validSignature,
			wantErr:       false,
		},
		{
			name:         "empty hash",
			documentHash: []byte(""),
			signatureData: validSignature,
			wantErr:      true,
			errContains:  "document hash cannot be empty",
		},
		{
			name:          "nil signature data",
			documentHash:  testHash,
			signatureData: nil,
			wantErr:       true,
			errContains:   "signature data cannot be nil",
		},
		{
			name:         "empty signature",
			documentHash: testHash,
			signatureData: &SignatureData{
				Signature: []byte(""),
				Hash:      testHash,
				Algorithm: "RSA-PSS-SHA256",
			},
			wantErr:     true,
			errContains: "signature cannot be empty",
		},
		{
			name:         "invalid signature",
			documentHash: testHash,
			signatureData: &SignatureData{
				Signature: []byte("invalid_signature"),
				Hash:      testHash,
				Algorithm: "RSA-PSS-SHA256",
			},
			wantErr:     true,
			errContains: "signature verification failed",
		},
		{
			name:         "wrong hash",
			documentHash: service.CalculateDocumentHash([]byte("different content")),
			signatureData: validSignature,
			wantErr:      true,
			errContains:  "signature verification failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.VerifySignature(tt.documentHash, tt.signatureData)

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

func TestSignatureService_IsSignatureValid(t *testing.T) {
	service := createTestSignatureService(t)
	
	// Create a proper SHA-256 hash for testing
	testData := []byte("test document content")
	testHash := service.CalculateDocumentHash(testData)

	// Create a valid signature for testing
	validSignature, err := service.SignDocument(testHash)
	require.NoError(t, err)

	tests := []struct {
		name          string
		documentHash  []byte
		signatureData *SignatureData
		expected      bool
	}{
		{
			name:          "valid signature",
			documentHash:  testHash,
			signatureData: validSignature,
			expected:      true,
		},
		{
			name:          "invalid signature",
			documentHash:  testHash,
			signatureData: &SignatureData{Signature: []byte("invalid"), Hash: testHash},
			expected:      false,
		},
		{
			name:          "nil signature data",
			documentHash:  testHash,
			signatureData: nil,
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.IsSignatureValid(tt.documentHash, tt.signatureData)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSignatureService_GetPublicKey(t *testing.T) {
	service := createTestSignatureService(t)

	publicKey := service.GetPublicKey()
	assert.NotNil(t, publicKey)
	assert.IsType(t, &rsa.PublicKey{}, publicKey)
}

func TestSignAndVerifyIntegration(t *testing.T) {
	service := createTestSignatureService(t)

	// Test data
	documentData := []byte("This is a test PDF document content for digital signature testing.")
	
	// Calculate hash
	documentHash := service.CalculateDocumentHash(documentData)
	assert.Len(t, documentHash, 32)

	// Sign the document
	signatureData, err := service.SignDocument(documentHash)
	require.NoError(t, err)
	assert.NotNil(t, signatureData)

	// Verify the signature
	err = service.VerifySignature(documentHash, signatureData)
	assert.NoError(t, err)

	// Test with different data should fail
	differentData := []byte("This is different content.")
	differentHash := service.CalculateDocumentHash(differentData)
	err = service.VerifySignature(differentHash, signatureData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signature verification failed")

	// Test IsSignatureValid convenience method
	assert.True(t, service.IsSignatureValid(documentHash, signatureData))
	assert.False(t, service.IsSignatureValid(differentHash, signatureData))
}

func TestLoadPrivateKey(t *testing.T) {
	tests := []struct {
		name        string
		keyPath     string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid private key",
			keyPath: "../../../../private_key.pem",
			wantErr: false,
		},
		{
			name:        "nonexistent file",
			keyPath:     "nonexistent.pem",
			wantErr:     true,
			errContains: "failed to read private key file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := loadPrivateKey(tt.keyPath)

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

func TestLoadPublicKey(t *testing.T) {
	tests := []struct {
		name        string
		keyPath     string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid public key",
			keyPath: "../../../../public_key.pem",
			wantErr: false,
		},
		{
			name:        "nonexistent file",
			keyPath:     "nonexistent.pem",
			wantErr:     true,
			errContains: "failed to read public key file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := loadPublicKey(tt.keyPath)

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

func TestInvalidPEMFormats(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		content     string
		keyType     string
		errContains string
	}{
		{
			name:        "invalid PEM format",
			content:     "not a pem file",
			keyType:     "private",
			errContains: "failed to decode PEM block",
		},
		{
			name: "wrong key type in private key",
			content: `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAu9W3f4+/TS7e/O363jhO
-----END PUBLIC KEY-----`,
			keyType:     "private",
			errContains: "failed to parse private key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			filePath := filepath.Join(tempDir, "test_key.pem")
			err := os.WriteFile(filePath, []byte(tt.content), 0600)
			require.NoError(t, err)

			if tt.keyType == "private" {
				key, err := loadPrivateKey(filePath)
				assert.Error(t, err)
				assert.Nil(t, key)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				key, err := loadPublicKey(filePath)
				assert.Error(t, err)
				assert.Nil(t, key)
				assert.Contains(t, err.Error(), tt.errContains)
			}
		})
	}
}

// createTestSignatureService creates a signature service for testing
func createTestSignatureService(t *testing.T) *SignatureService {
	service, err := NewSignatureService("../../../../private_key.pem", "../../../../public_key.pem")
	require.NoError(t, err)
	return service
}

// createTestKeyPair creates a temporary RSA key pair for testing
func createTestKeyPair(t *testing.T) (string, string) {
	// Generate test key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Create temporary directory
	tempDir := t.TempDir()

	// Save private key
	privateKeyPath := filepath.Join(tempDir, "private_key.pem")
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	err = os.WriteFile(privateKeyPath, privateKeyPEM, 0600)
	require.NoError(t, err)

	// Save public key
	publicKeyPath := filepath.Join(tempDir, "public_key.pem")
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	err = os.WriteFile(publicKeyPath, publicKeyPEM, 0644)
	require.NoError(t, err)

	return privateKeyPath, publicKeyPath
}

func TestNewSignatureServiceFromKeyManager(t *testing.T) {
	tests := []struct {
		name        string
		setupKM     func() *KeyManager
		wantErr     bool
		errContains string
	}{
		{
			name: "valid key manager",
			setupKM: func() *KeyManager {
				km, err := NewKeyManagerFromFiles("../../../../private_key.pem", "../../../../public_key.pem")
				require.NoError(t, err)
				return km
			},
			wantErr: false,
		},
		{
			name: "nil key manager",
			setupKM: func() *KeyManager {
				return nil
			},
			wantErr:     true,
			errContains: "key manager cannot be nil",
		},
		{
			name: "invalid key manager",
			setupKM: func() *KeyManager {
				// Create a key manager with invalid keys
				privateKey1, _ := rsa.GenerateKey(rand.Reader, 2048)
				privateKey2, _ := rsa.GenerateKey(rand.Reader, 2048)
				return &KeyManager{
					privateKey: privateKey1,
					publicKey:  &privateKey2.PublicKey, // Mismatched keys
					keyID:      "test",
					createdAt:  time.Now(),
				}
			},
			wantErr:     true,
			errContains: "key validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			km := tt.setupKM()
			service, err := NewSignatureServiceFromKeyManager(km)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, service)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
				assert.NotNil(t, service.GetPublicKey())
			}
		})
	}
}

func TestKeyManagerSignatureServiceIntegration(t *testing.T) {
	// Create key manager
	km, err := NewKeyManagerFromFiles("../../../../private_key.pem", "../../../../public_key.pem")
	require.NoError(t, err)

	// Create signature service from key manager
	service, err := NewSignatureServiceFromKeyManager(km)
	require.NoError(t, err)

	// Test document signing and verification
	document := []byte("Integration test document for KeyManager and SignatureService")
	hash := service.CalculateDocumentHash(document)

	// Sign document
	signature, err := service.SignDocument(hash)
	require.NoError(t, err)

	// Verify signature
	err = service.VerifySignature(hash, signature)
	assert.NoError(t, err)

	// Test that the public keys match
	assert.Equal(t, km.GetPublicKey(), service.GetPublicKey())

	// Test key rotation integration
	newKeyPair, err := km.GenerateNewKeyPair(2048)
	require.NoError(t, err)

	err = km.RotateKeys(newKeyPair)
	require.NoError(t, err)

	// Create new signature service with rotated keys
	newService, err := NewSignatureServiceFromKeyManager(km)
	require.NoError(t, err)

	// Old signature should not verify with new keys
	err = newService.VerifySignature(hash, signature)
	assert.Error(t, err, "Old signature should not verify with new keys")

	// New signature should work with new keys
	newSignature, err := newService.SignDocument(hash)
	require.NoError(t, err)

	err = newService.VerifySignature(hash, newSignature)
	assert.NoError(t, err, "New signature should verify with new keys")
}