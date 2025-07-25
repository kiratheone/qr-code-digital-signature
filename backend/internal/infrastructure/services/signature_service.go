package services

import (
	"digital-signature-system/internal/config"
	"digital-signature-system/internal/domain/services"
)

// NewSignatureService creates a new signature service with keys from config
func NewSignatureService(cfg *config.Config) (services.SignatureService, error) {
	return services.NewSignatureService(cfg.Security.PrivateKey, cfg.Security.PublicKey)
}