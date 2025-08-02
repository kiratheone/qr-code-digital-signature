package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"
)

// KeyManager handles secure key storage and management
type KeyManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	keyID      string
	createdAt  time.Time
}

// KeyPair represents an RSA key pair with metadata
type KeyPair struct {
	PrivateKey string    `json:"private_key"`
	PublicKey  string    `json:"public_key"`
	KeyID      string    `json:"key_id"`
	CreatedAt  time.Time `json:"created_at"`
	Algorithm  string    `json:"algorithm"`
}

// NewKeyManager creates a new key manager from environment variables
func NewKeyManager() (*KeyManager, error) {
	// Try to load keys from environment variables first
	// Support both RSA_PRIVATE_KEY and PRIVATE_KEY for backward compatibility
	privateKeyEnv := os.Getenv("RSA_PRIVATE_KEY")
	if privateKeyEnv == "" {
		privateKeyEnv = os.Getenv("PRIVATE_KEY")
	}
	
	if privateKeyEnv != "" {
		publicKeyEnv := os.Getenv("RSA_PUBLIC_KEY")
		if publicKeyEnv == "" {
			publicKeyEnv = os.Getenv("PUBLIC_KEY")
		}
		return NewKeyManagerFromEnv(privateKeyEnv, publicKeyEnv)
	}

	// Fallback to file-based keys
	privateKeyPath := os.Getenv("PRIVATE_KEY_PATH")
	if privateKeyPath == "" {
		privateKeyPath = "private_key.pem"
	}

	publicKeyPath := os.Getenv("PUBLIC_KEY_PATH")
	if publicKeyPath == "" {
		publicKeyPath = "public_key.pem"
	}

	return NewKeyManagerFromFiles(privateKeyPath, publicKeyPath)
}

// NewKeyManagerFromEnv creates a key manager from environment variables
func NewKeyManagerFromEnv(privateKeyEnv, publicKeyEnv string) (*KeyManager, error) {
	if privateKeyEnv == "" {
		return nil, fmt.Errorf("private key environment variable is empty")
	}

	// Decode base64 if needed
	privateKeyPEM := privateKeyEnv
	if !strings.Contains(privateKeyPEM, "-----BEGIN") {
		decoded, err := base64.StdEncoding.DecodeString(privateKeyEnv)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 private key: %w", err)
		}
		privateKeyPEM = string(decoded)
	}

	privateKey, err := parsePrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key from environment: %w", err)
	}

	var publicKey *rsa.PublicKey
	if publicKeyEnv != "" {
		publicKeyPEM := publicKeyEnv
		if !strings.Contains(publicKeyPEM, "-----BEGIN") {
			decoded, err := base64.StdEncoding.DecodeString(publicKeyEnv)
			if err != nil {
				return nil, fmt.Errorf("failed to decode base64 public key: %w", err)
			}
			publicKeyPEM = string(decoded)
		}

		publicKey, err = parsePublicKeyFromPEM(publicKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key from environment: %w", err)
		}
	} else {
		// Derive public key from private key
		publicKey = &privateKey.PublicKey
	}

	keyID := generateKeyID(publicKey)

	km := &KeyManager{
		privateKey: privateKey,
		publicKey:  publicKey,
		keyID:      keyID,
		createdAt:  time.Now(),
	}

	// Validate keys during creation
	if err := km.ValidateKeys(); err != nil {
		return nil, fmt.Errorf("key validation failed: %w", err)
	}

	return km, nil
}

// NewKeyManagerFromFiles creates a key manager from PEM files
func NewKeyManagerFromFiles(privateKeyPath, publicKeyPath string) (*KeyManager, error) {
	privateKey, err := loadPrivateKeyFromFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key from file: %w", err)
	}

	publicKey, err := loadPublicKeyFromFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key from file: %w", err)
	}

	keyID := generateKeyID(publicKey)

	km := &KeyManager{
		privateKey: privateKey,
		publicKey:  publicKey,
		keyID:      keyID,
		createdAt:  time.Now(),
	}

	// Validate keys during creation
	if err := km.ValidateKeys(); err != nil {
		return nil, fmt.Errorf("key validation failed: %w", err)
	}

	return km, nil
}

// GetPrivateKey returns the private key
func (km *KeyManager) GetPrivateKey() *rsa.PrivateKey {
	return km.privateKey
}

// GetPublicKey returns the public key
func (km *KeyManager) GetPublicKey() *rsa.PublicKey {
	return km.publicKey
}

// GetKeyID returns the key identifier
func (km *KeyManager) GetKeyID() string {
	return km.keyID
}

// GetCreatedAt returns when the key was loaded/created
func (km *KeyManager) GetCreatedAt() time.Time {
	return km.createdAt
}

// GetKeySize returns the key size in bits
func (km *KeyManager) GetKeySize() int {
	if km.privateKey == nil {
		return 0
	}
	return km.privateKey.N.BitLen()
}

// ShouldRotateKey checks if the key should be rotated based on age
func (km *KeyManager) ShouldRotateKey(maxAge time.Duration) bool {
	return time.Since(km.createdAt) > maxAge
}

// GetKeyAge returns the age of the current key
func (km *KeyManager) GetKeyAge() time.Duration {
	return time.Since(km.createdAt)
}

// ClearKeys securely clears the keys from memory
func (km *KeyManager) ClearKeys() {
	if km.privateKey != nil {
		// Zero out the private key components
		if km.privateKey.D != nil {
			km.privateKey.D.SetInt64(0)
		}
		if len(km.privateKey.Primes) > 0 {
			for _, prime := range km.privateKey.Primes {
				if prime != nil {
					prime.SetInt64(0)
				}
			}
		}
		km.privateKey = nil
	}
	km.publicKey = nil
	km.keyID = ""
}

// ValidateKeys validates that the key pair is valid and matches
func (km *KeyManager) ValidateKeys() error {
	if km.privateKey == nil {
		return fmt.Errorf("private key is nil")
	}

	if km.publicKey == nil {
		return fmt.Errorf("public key is nil")
	}

	// Validate key size (minimum 2048 bits for security)
	keySize := km.privateKey.N.BitLen()
	if keySize < 2048 {
		return fmt.Errorf("key size %d bits is below minimum security requirement of 2048 bits", keySize)
	}

	// Verify that the public key matches the private key
	if km.privateKey.PublicKey.N.Cmp(km.publicKey.N) != 0 {
		return fmt.Errorf("public key does not match private key")
	}

	if km.privateKey.PublicKey.E != km.publicKey.E {
		return fmt.Errorf("public key exponent does not match private key")
	}

	// Test key pair by signing and verifying a test message
	testMessage := []byte("key validation test")
	hash := sha256.Sum256(testMessage)
	signature, err := rsa.SignPSS(rand.Reader, km.privateKey, crypto.SHA256, hash[:], nil)
	if err != nil {
		return fmt.Errorf("failed to sign test message: %w", err)
	}

	err = rsa.VerifyPSS(km.publicKey, crypto.SHA256, hash[:], signature, nil)
	if err != nil {
		return fmt.Errorf("failed to verify test signature: %w", err)
	}

	return nil
}

// GenerateNewKeyPair generates a new RSA key pair for rotation
func (km *KeyManager) GenerateNewKeyPair(keySize int) (*KeyPair, error) {
	if keySize < 2048 {
		return nil, fmt.Errorf("key size must be at least 2048 bits")
	}

	// Generate new private key
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Convert to PEM format
	privateKeyPEM, err := privateKeyToPEM(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert private key to PEM: %w", err)
	}

	publicKeyPEM, err := publicKeyToPEM(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert public key to PEM: %w", err)
	}

	keyID := generateKeyID(&privateKey.PublicKey)

	return &KeyPair{
		PrivateKey: privateKeyPEM,
		PublicKey:  publicKeyPEM,
		KeyID:      keyID,
		CreatedAt:  time.Now(),
		Algorithm:  "RSA-2048",
	}, nil
}

// RotateKeys replaces the current keys with new ones
func (km *KeyManager) RotateKeys(newKeyPair *KeyPair) error {
	if newKeyPair == nil {
		return fmt.Errorf("new key pair cannot be nil")
	}

	// Parse the new keys
	privateKey, err := parsePrivateKeyFromPEM(newKeyPair.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to parse new private key: %w", err)
	}

	publicKey, err := parsePublicKeyFromPEM(newKeyPair.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to parse new public key: %w", err)
	}

	// Validate the new key pair
	tempManager := &KeyManager{
		privateKey: privateKey,
		publicKey:  publicKey,
		keyID:      newKeyPair.KeyID,
		createdAt:  newKeyPair.CreatedAt,
	}

	if err := tempManager.ValidateKeys(); err != nil {
		return fmt.Errorf("new key pair validation failed: %w", err)
	}

	// Replace current keys
	km.privateKey = privateKey
	km.publicKey = publicKey
	km.keyID = newKeyPair.KeyID
	km.createdAt = newKeyPair.CreatedAt

	return nil
}

// ExportKeyPairForStorage exports the current key pair for secure storage
func (km *KeyManager) ExportKeyPairForStorage() (*KeyPair, error) {
	privateKeyPEM, err := privateKeyToPEM(km.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to export private key: %w", err)
	}

	publicKeyPEM, err := publicKeyToPEM(km.publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to export public key: %w", err)
	}

	return &KeyPair{
		PrivateKey: privateKeyPEM,
		PublicKey:  publicKeyPEM,
		KeyID:      km.keyID,
		CreatedAt:  km.createdAt,
		Algorithm:  "RSA-2048",
	}, nil
}

// Helper functions

func parsePrivateKeyFromPEM(pemData string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	// Try PKCS#8 format first
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
		return nil, fmt.Errorf("key is not an RSA private key")
	}

	// Try PKCS#1 format
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	return nil, fmt.Errorf("failed to parse private key")
}

func parsePublicKeyFromPEM(pemData string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("key is not an RSA public key")
	}

	return rsaKey, nil
}

func loadPrivateKeyFromFile(keyPath string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	return parsePrivateKeyFromPEM(string(keyData))
}

func loadPublicKeyFromFile(keyPath string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	return parsePublicKeyFromPEM(string(keyData))
}

func privateKeyToPEM(privateKey *rsa.PrivateKey) (string, error) {
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal private key: %w", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	return string(privateKeyPEM), nil
}

func publicKeyToPEM(publicKey *rsa.PublicKey) (string, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(publicKeyPEM), nil
}

func generateKeyID(publicKey *rsa.PublicKey) string {
	// Generate a simple key ID based on the public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return fmt.Sprintf("key_%d", time.Now().Unix())
	}

	// Use SHA-256 hash of the public key and take first 16 bytes for better uniqueness
	hash := sha256.Sum256(publicKeyBytes)
	return fmt.Sprintf("key_%x", hash[:16])
}