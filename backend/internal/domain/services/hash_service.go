package services

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
)

// HashService defines the interface for hash operations
type HashService interface {
	// CalculateHash calculates the hash of a document
	CalculateHash(document []byte) ([]byte, error)
	
	// CalculateHashFromReader calculates the hash from a reader
	CalculateHashFromReader(reader io.Reader) ([]byte, error)
	
	// VerifyHash verifies if a document matches a hash
	VerifyHash(document []byte, hash []byte) (bool, error)
	
	// HashToString converts a hash to a string representation
	HashToString(hash []byte) string
	
	// StringToHash converts a string representation to a hash
	StringToHash(hashStr string) ([]byte, error)
}

// SHA256HashService implements HashService using SHA-256
type SHA256HashService struct{}

// NewSHA256HashService creates a new SHA-256 hash service
func NewSHA256HashService() *SHA256HashService {
	return &SHA256HashService{}
}

// CalculateHash calculates the SHA-256 hash of a document
func (s *SHA256HashService) CalculateHash(document []byte) ([]byte, error) {
	hash := sha256.Sum256(document)
	return hash[:], nil
}

// CalculateHashFromReader calculates the SHA-256 hash from a reader
func (s *SHA256HashService) CalculateHashFromReader(reader io.Reader) ([]byte, error) {
	h := sha256.New()
	if _, err := io.Copy(h, reader); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

// VerifyHash verifies if a document matches a hash
func (s *SHA256HashService) VerifyHash(document []byte, hash []byte) (bool, error) {
	docHash, err := s.CalculateHash(document)
	if err != nil {
		return false, err
	}
	
	// Compare the hashes
	return compareHashes(docHash, hash), nil
}

// HashToString converts a hash to a string representation
func (s *SHA256HashService) HashToString(hash []byte) string {
	return hex.EncodeToString(hash)
}

// StringToHash converts a string representation to a hash
func (s *SHA256HashService) StringToHash(hashStr string) ([]byte, error) {
	return hex.DecodeString(hashStr)
}

// Helper function to compare hashes
func compareHashes(hash1, hash2 []byte) bool {
	if len(hash1) != len(hash2) {
		return false
	}
	
	// Constant-time comparison to prevent timing attacks
	var result byte
	for i := 0; i < len(hash1); i++ {
		result |= hash1[i] ^ hash2[i]
	}
	
	return result == 0
}