package services

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
)

// SignatureService defines the interface for digital signature operations
type SignatureService interface {
	// SignDocument signs a document hash with the private key
	SignDocument(docHash []byte) ([]byte, error)
	
	// VerifySignature verifies a signature against a document hash
	VerifySignature(docHash []byte, signature []byte) (bool, error)
	
	// GenerateKeyPair generates a new RSA key pair
	GenerateKeyPair(bits int) (privateKey string, publicKey string, err error)
}

// RSASignatureService implements SignatureService using RSA
type RSASignatureService struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// NewRSASignatureService creates a new RSA signature service
func NewRSASignatureService(privateKeyPEM, publicKeyPEM string) (*RSASignatureService, error) {
	privateKey, err := parsePrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKey, err := parsePublicKey(publicKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return &RSASignatureService{
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

// SignDocument signs a document hash with the private key
func (s *RSASignatureService) SignDocument(docHash []byte) ([]byte, error) {
	if s.privateKey == nil {
		return nil, errors.New("private key not available")
	}

	// Sign the hash with the private key
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, docHash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign document: %w", err)
	}

	return signature, nil
}

// VerifySignature verifies a signature against a document hash
func (s *RSASignatureService) VerifySignature(docHash []byte, signature []byte) (bool, error) {
	if s.publicKey == nil {
		return false, errors.New("public key not available")
	}

	// Verify the signature
	err := rsa.VerifyPKCS1v15(s.publicKey, crypto.SHA256, docHash, signature)
	if err != nil {
		if errors.Is(err, rsa.ErrVerification) {
			return false, nil
		}
		return false, fmt.Errorf("failed to verify signature: %w", err)
	}

	return true, nil
}

// GenerateKeyPair generates a new RSA key pair
func (s *RSASignatureService) GenerateKeyPair(bits int) (string, string, error) {
	// Generate a new RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Convert private key to PEM format
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Convert public key to PEM format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal public key: %w", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(privateKeyPEM), string(publicKeyPEM), nil
}

// CalculateHash calculates the SHA-256 hash of a document
func CalculateHash(document []byte) []byte {
	hash := sha256.Sum256(document)
	return hash[:]
}

// EncodeSignature encodes a signature as a base64 string
func EncodeSignature(signature []byte) string {
	return base64.StdEncoding.EncodeToString(signature)
}

// DecodeSignature decodes a base64 signature string
func DecodeSignature(encodedSignature string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encodedSignature)
}

// Helper functions for parsing keys
func parsePrivateKey(privateKeyPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return privateKey, nil
}

func parsePublicKey(publicKeyPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	publicKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return publicKey, nil
}