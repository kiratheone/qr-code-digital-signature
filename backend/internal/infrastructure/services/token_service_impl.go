package services

import (
	"digital-signature-system/internal/config"
	"digital-signature-system/internal/domain/services"
)

// NewTokenService creates a new token service
func NewTokenService(cfg *config.Config) services.TokenService {
	return services.NewJWTTokenService(cfg.Security.JWTSecret, "digital-signature-system")
}