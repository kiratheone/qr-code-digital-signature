package services

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// PasswordService defines the interface for password operations
type PasswordService interface {
	// HashPassword hashes a password
	HashPassword(password string) (string, error)
	
	// VerifyPassword verifies a password against a hash
	VerifyPassword(password, hash string) (bool, error)
}

// BCryptPasswordService implements PasswordService using bcrypt
type BCryptPasswordService struct {
	cost int
}

// NewBCryptPasswordService creates a new bcrypt password service
func NewBCryptPasswordService(cost int) *BCryptPasswordService {
	// If cost is invalid, use default cost
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	
	return &BCryptPasswordService{
		cost: cost,
	}
}

// HashPassword hashes a password using bcrypt
func (s *BCryptPasswordService) HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}
	
	// Generate hash from password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	
	return string(hash), nil
}

// VerifyPassword verifies a password against a hash
func (s *BCryptPasswordService) VerifyPassword(password, hash string) (bool, error) {
	if password == "" || hash == "" {
		return false, errors.New("password or hash cannot be empty")
	}
	
	// Compare password with hash
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, fmt.Errorf("failed to verify password: %w", err)
	}
	
	return true, nil
}