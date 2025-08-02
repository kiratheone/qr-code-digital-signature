package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// SignatureService handles RSA digital signature operations
type SignatureService struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// SignatureData represents a digital signature
type SignatureData struct {
	Signature []byte `json:"signature"`
	Hash      []byte `json:"hash"`
	Algorithm string `json:"algorithm"`
}

// NewSignatureService creates a new signature service with RSA keys
func NewSignatureService(privateKeyPath, publicKeyPath string) (*SignatureService, error) {
	privateKey, err := loadPrivateKey(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	publicKey, err := loadPublicKey(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key: %w", err)
	}

	return &SignatureService{
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

// NewSignatureServiceFromKeyManager creates a new signature service from a KeyManager
func NewSignatureServiceFromKeyManager(km *KeyManager) (*SignatureService, error) {
	if km == nil {
		return nil, fmt.Errorf("key manager cannot be nil")
	}

	// Validate keys before using them
	if err := km.ValidateKeys(); err != nil {
		return nil, fmt.Errorf("key validation failed: %w", err)
	}

	return &SignatureService{
		privateKey: km.GetPrivateKey(),
		publicKey:  km.GetPublicKey(),
	}, nil
}

// CalculateDocumentHash calculates SHA-256 hash of document data
func (s *SignatureService) CalculateDocumentHash(documentData []byte) []byte {
	hash := sha256.Sum256(documentData)
	return hash[:]
}

// SignDocument creates a digital signature for the document hash
func (s *SignatureService) SignDocument(documentHash []byte) (*SignatureData, error) {
	if len(documentHash) == 0 {
		return nil, fmt.Errorf("document hash cannot be empty")
	}

	// Sign the hash using RSA-PSS with SHA-256
	signature, err := rsa.SignPSS(rand.Reader, s.privateKey, crypto.SHA256, documentHash, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to sign document hash: %w", err)
	}

	return &SignatureData{
		Signature: signature,
		Hash:      documentHash,
		Algorithm: "RSA-PSS-SHA256",
	}, nil
}

// VerifySignature verifies a digital signature against the document hash
func (s *SignatureService) VerifySignature(documentHash []byte, signatureData *SignatureData) error {
	if len(documentHash) == 0 {
		return fmt.Errorf("document hash cannot be empty")
	}

	if signatureData == nil {
		return fmt.Errorf("signature data cannot be nil")
	}

	if len(signatureData.Signature) == 0 {
		return fmt.Errorf("signature cannot be empty")
	}

	// Verify the signature using RSA-PSS with SHA-256
	err := rsa.VerifyPSS(s.publicKey, crypto.SHA256, documentHash, signatureData.Signature, nil)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

// IsSignatureValid checks if a signature is valid for the given document hash
func (s *SignatureService) IsSignatureValid(documentHash []byte, signatureData *SignatureData) bool {
	return s.VerifySignature(documentHash, signatureData) == nil
}

// GetPublicKey returns the public key for external verification
func (s *SignatureService) GetPublicKey() *rsa.PublicKey {
	return s.publicKey
}

// loadPrivateKey loads RSA private key from PEM file
func loadPrivateKey(keyPath string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
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

// loadPublicKey loads RSA public key from PEM file
func loadPublicKey(keyPath string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}

	// Parse PKIX public key
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