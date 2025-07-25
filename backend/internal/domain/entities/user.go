package entities

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Username     string    `json:"username" gorm:"uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	FullName     string    `json:"full_name" gorm:"not null"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	Role         string    `json:"role" gorm:"default:user"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
}

func NewUser(username, passwordHash, fullName, email string) *User {
	return &User{
		ID:           uuid.New().String(),
		Username:     username,
		PasswordHash: passwordHash,
		FullName:     fullName,
		Email:        email,
		Role:         "user",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}