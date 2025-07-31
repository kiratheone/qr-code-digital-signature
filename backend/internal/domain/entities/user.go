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

func (u *User) BeforeCreate() error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}