package services

import (
	"digital-signature-system/internal/config"
	"digital-signature-system/internal/domain/services"
	"fmt"
	"os"
	"path/filepath"
)

// KeyServiceFactory creates a new key service based on configuration
func NewKeyService(cfg *config.Config) (services.KeyService, error) {
	// Check if environment variables are set
	privateKeyEnv := "PRIVATE_KEY"
	publicKeyEnv := "PUBLIC_KEY"
	
	if os.Getenv(privateKeyEnv) != "" && os.Getenv(publicKeyEnv) != "" {
		return services.NewEnvironmentKeyService(privateKeyEnv, publicKeyEnv), nil
	}
	
	// If environment variables are not set, use file-based key service
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	
	keyDir := filepath.Join(homeDir, ".digital-signature-system", "keys")
	privateKeyPath := filepath.Join(keyDir, "private.pem")
	publicKeyPath := filepath.Join(keyDir, "public.pem")
	
	// Create key directory if it doesn't exist
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create key directory: %w", err)
	}
	
	return services.NewFileKeyService(privateKeyPath, publicKeyPath), nil
}

// LoadOrGenerateKeys loads existing keys or generates new ones if they don't exist
func LoadOrGenerateKeys(keyService services.KeyService) (privateKey string, publicKey string, err error) {
	// Try to load existing keys
	privateKey, publicKey, err = keyService.LoadKeys()
	if err == nil {
		// Validate keys
		if err := services.ValidateRSAPrivateKey(privateKey); err != nil {
			return "", "", fmt.Errorf("invalid private key: %w", err)
		}
		if err := services.ValidateRSAPublicKey(publicKey); err != nil {
			return "", "", fmt.Errorf("invalid public key: %w", err)
		}
		return privateKey, publicKey, nil
	}
	
	// If keys don't exist, generate new ones
	fmt.Println("Keys not found or invalid, generating new RSA key pair...")
	return keyService.GenerateAndSaveKeys(2048)
}