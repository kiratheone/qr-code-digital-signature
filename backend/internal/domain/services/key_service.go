package services

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// KeyService defines the interface for key management operations
type KeyService interface {
	// LoadKeys loads the RSA keys from the specified sources
	LoadKeys() (privateKey string, publicKey string, err error)
	
	// GenerateAndSaveKeys generates a new RSA key pair and saves them
	GenerateAndSaveKeys(bits int) (privateKey string, publicKey string, err error)
	
	// RotateKeys rotates the RSA keys
	RotateKeys(bits int) (privateKey string, publicKey string, err error)
}

// FileKeyService implements KeyService using file storage
type FileKeyService struct {
	privateKeyPath string
	publicKeyPath  string
}

// NewFileKeyService creates a new file-based key service
func NewFileKeyService(privateKeyPath, publicKeyPath string) *FileKeyService {
	return &FileKeyService{
		privateKeyPath: privateKeyPath,
		publicKeyPath:  publicKeyPath,
	}
}

// LoadKeys loads the RSA keys from files
func (s *FileKeyService) LoadKeys() (string, string, error) {
	// Check if key files exist
	if _, err := os.Stat(s.privateKeyPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("private key file not found: %s", s.privateKeyPath)
	}
	if _, err := os.Stat(s.publicKeyPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("public key file not found: %s", s.publicKeyPath)
	}

	// Read private key
	privateKeyBytes, err := ioutil.ReadFile(s.privateKeyPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read private key: %w", err)
	}

	// Read public key
	publicKeyBytes, err := ioutil.ReadFile(s.publicKeyPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read public key: %w", err)
	}

	return string(privateKeyBytes), string(publicKeyBytes), nil
}

// GenerateAndSaveKeys generates a new RSA key pair and saves them to files
func (s *FileKeyService) GenerateAndSaveKeys(bits int) (string, string, error) {
	// Create directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(s.privateKeyPath), 0700); err != nil {
		return "", "", fmt.Errorf("failed to create directory for private key: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.publicKeyPath), 0700); err != nil {
		return "", "", fmt.Errorf("failed to create directory for public key: %w", err)
	}

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

	// Write private key to file with restricted permissions
	if err := ioutil.WriteFile(s.privateKeyPath, privateKeyPEM, 0600); err != nil {
		return "", "", fmt.Errorf("failed to write private key: %w", err)
	}

	// Write public key to file
	if err := ioutil.WriteFile(s.publicKeyPath, publicKeyPEM, 0644); err != nil {
		return "", "", fmt.Errorf("failed to write public key: %w", err)
	}

	return string(privateKeyPEM), string(publicKeyPEM), nil
}

// RotateKeys rotates the RSA keys
func (s *FileKeyService) RotateKeys(bits int) (string, string, error) {
	// Backup existing keys if they exist
	if _, err := os.Stat(s.privateKeyPath); err == nil {
		backupTime := time.Now().Format("20060102-150405")
		privateKeyBackupPath := s.privateKeyPath + "." + backupTime + ".bak"
		if err := copyFile(s.privateKeyPath, privateKeyBackupPath); err != nil {
			return "", "", fmt.Errorf("failed to backup private key: %w", err)
		}
	}

	if _, err := os.Stat(s.publicKeyPath); err == nil {
		backupTime := time.Now().Format("20060102-150405")
		publicKeyBackupPath := s.publicKeyPath + "." + backupTime + ".bak"
		if err := copyFile(s.publicKeyPath, publicKeyBackupPath); err != nil {
			return "", "", fmt.Errorf("failed to backup public key: %w", err)
		}
	}

	// Generate and save new keys
	return s.GenerateAndSaveKeys(bits)
}

// EnvironmentKeyService implements KeyService using environment variables
type EnvironmentKeyService struct {
	privateKeyEnv string
	publicKeyEnv  string
}

// NewEnvironmentKeyService creates a new environment-based key service
func NewEnvironmentKeyService(privateKeyEnv, publicKeyEnv string) *EnvironmentKeyService {
	return &EnvironmentKeyService{
		privateKeyEnv: privateKeyEnv,
		publicKeyEnv:  publicKeyEnv,
	}
}

// LoadKeys loads the RSA keys from environment variables
func (s *EnvironmentKeyService) LoadKeys() (string, string, error) {
	privateKey := os.Getenv(s.privateKeyEnv)
	if privateKey == "" {
		return "", "", fmt.Errorf("private key environment variable not set: %s", s.privateKeyEnv)
	}

	publicKey := os.Getenv(s.publicKeyEnv)
	if publicKey == "" {
		return "", "", fmt.Errorf("public key environment variable not set: %s", s.publicKeyEnv)
	}

	return privateKey, publicKey, nil
}

// GenerateAndSaveKeys generates a new RSA key pair and saves them to environment variables
func (s *EnvironmentKeyService) GenerateAndSaveKeys(bits int) (string, string, error) {
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

	// Set environment variables
	if err := os.Setenv(s.privateKeyEnv, string(privateKeyPEM)); err != nil {
		return "", "", fmt.Errorf("failed to set private key environment variable: %w", err)
	}

	if err := os.Setenv(s.publicKeyEnv, string(publicKeyPEM)); err != nil {
		return "", "", fmt.Errorf("failed to set public key environment variable: %w", err)
	}

	return string(privateKeyPEM), string(publicKeyPEM), nil
}

// RotateKeys rotates the RSA keys
func (s *EnvironmentKeyService) RotateKeys(bits int) (string, string, error) {
	// No need to backup environment variables, just generate new keys
	return s.GenerateAndSaveKeys(bits)
}

// Helper function to copy a file
func copyFile(src, dst string) error {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, data, 0600)
}

// ValidateRSAPrivateKey validates an RSA private key
func ValidateRSAPrivateKey(privateKeyPEM string) error {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return errors.New("failed to decode PEM block containing private key")
	}

	_, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	return nil
}

// ValidateRSAPublicKey validates an RSA public key
func ValidateRSAPublicKey(publicKeyPEM string) error {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil || block.Type != "PUBLIC KEY" {
		return errors.New("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	_, ok := pub.(*rsa.PublicKey)
	if !ok {
		return errors.New("not an RSA public key")
	}

	return nil
}