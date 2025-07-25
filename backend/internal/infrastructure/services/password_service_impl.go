package services

import (
	"digital-signature-system/internal/domain/services"

	"golang.org/x/crypto/bcrypt"
)

// NewPasswordService creates a new password service
func NewPasswordService() services.PasswordService {
	return services.NewBCryptPasswordService(bcrypt.DefaultCost)
}