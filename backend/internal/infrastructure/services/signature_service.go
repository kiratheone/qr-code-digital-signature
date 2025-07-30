package services

import (
	"digital-signature-system/internal/config"
	"digital-signature-system/internal/domain/services"
	"fmt"
)

// NewSignatureService creates a new signature service with keys from config
func NewSignatureService(cfg *config.Config) (services.SignatureService, error) {
	privateKey := cfg.Security.PrivateKey
	publicKey := cfg.Security.PublicKey

	if privateKey == "" || publicKey == "" {
		// Generate a new key pair if not provided
		tempService := &services.RSASignatureService{}
		var err error
		privateKey, publicKey, err = tempService.GenerateKeyPair(2048)
		if err != nil {
			return nil, fmt.Errorf("failed to generate key pair: %w", err)
		}

		// Log warning that keys were generated
		fmt.Println("WARNING: RSA keys were not provided, generated new keys")
		fmt.Println("Private Key:")
		fmt.Println(privateKey)
		fmt.Println("Public Key:")
		fmt.Println(publicKey)
	}

	return services.NewRSASignatureService(privateKey, publicKey)
}