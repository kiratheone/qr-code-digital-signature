package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Session struct {
	ID           string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID       string    `json:"user_id" gorm:"not null;index:idx_sessions_user_id"`
	SessionToken string    `json:"session_token" gorm:"uniqueIndex:idx_sessions_token;not null"`
	RefreshToken string    `json:"refresh_token" gorm:"uniqueIndex:idx_sessions_refresh_token;not null"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	LastAccessed time.Time `json:"last_accessed"`
	User         User      `json:"user" gorm:"foreignKey:UserID"`
}

func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}