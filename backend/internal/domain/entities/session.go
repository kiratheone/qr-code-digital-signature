package entities

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID           string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID       string    `json:"user_id" gorm:"not null"`
	SessionToken string    `json:"session_token" gorm:"uniqueIndex;not null"`
	RefreshToken string    `json:"refresh_token" gorm:"uniqueIndex;not null"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	LastAccessed time.Time `json:"last_accessed"`
	User         User      `json:"user" gorm:"foreignKey:UserID"`
}

func NewSession(userID, sessionToken, refreshToken string, expiresAt time.Time) *Session {
	return &Session{
		ID:           uuid.New().String(),
		UserID:       userID,
		SessionToken: sessionToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
	}
}