package services

import (
	"digital-signature-system/internal/domain/services"
)

// NewHashService creates a new hash service
func NewHashService() services.HashService {
	return services.NewSHA256HashService()
}