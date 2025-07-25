package services

import (
	"digital-signature-system/internal/config"
	"digital-signature-system/internal/domain/services"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// KeyRotationService handles key rotation
type KeyRotationService struct {
	keyService      services.KeyService
	rotationDays    int
	lastRotationFile string
}

// NewKeyRotationService creates a new key rotation service
func NewKeyRotationService(cfg *config.Config, keyService services.KeyService) *KeyRotationService {
	// Get user home directory for storing rotation timestamp
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Warning: Failed to get user home directory: %v\n", err)
		homeDir = os.TempDir()
	}

	rotationDir := filepath.Join(homeDir, ".digital-signature-system", "rotation")
	if err := os.MkdirAll(rotationDir, 0700); err != nil {
		fmt.Printf("Warning: Failed to create rotation directory: %v\n", err)
	}

	return &KeyRotationService{
		keyService:      keyService,
		rotationDays:    cfg.Security.KeyRotationDays,
		lastRotationFile: filepath.Join(rotationDir, "last_rotation.txt"),
	}
}

// CheckAndRotateKeys checks if keys need to be rotated and rotates them if necessary
func (s *KeyRotationService) CheckAndRotateKeys() (bool, error) {
	// If rotation is disabled, do nothing
	if s.rotationDays <= 0 {
		return false, nil
	}

	// Check when keys were last rotated
	lastRotation, err := s.getLastRotationTime()
	if err != nil {
		return false, fmt.Errorf("failed to get last rotation time: %w", err)
	}

	// If keys have never been rotated, or it's time to rotate them
	now := time.Now()
	if lastRotation.IsZero() || now.Sub(lastRotation) > time.Duration(s.rotationDays)*24*time.Hour {
		// Rotate keys
		_, _, err := s.keyService.RotateKeys(2048)
		if err != nil {
			return false, fmt.Errorf("failed to rotate keys: %w", err)
		}

		// Update last rotation time
		if err := s.updateLastRotationTime(now); err != nil {
			return true, fmt.Errorf("keys rotated but failed to update last rotation time: %w", err)
		}

		return true, nil
	}

	return false, nil
}

// getLastRotationTime gets the time when keys were last rotated
func (s *KeyRotationService) getLastRotationTime() (time.Time, error) {
	// Check if last rotation file exists
	if _, err := os.Stat(s.lastRotationFile); os.IsNotExist(err) {
		return time.Time{}, nil
	}

	// Read last rotation time from file
	data, err := os.ReadFile(s.lastRotationFile)
	if err != nil {
		return time.Time{}, err
	}

	// Parse time
	lastRotation, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		return time.Time{}, err
	}

	return lastRotation, nil
}

// updateLastRotationTime updates the time when keys were last rotated
func (s *KeyRotationService) updateLastRotationTime(t time.Time) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(s.lastRotationFile), 0700); err != nil {
		return err
	}

	// Write time to file
	return os.WriteFile(s.lastRotationFile, []byte(t.Format(time.RFC3339)), 0600)
}