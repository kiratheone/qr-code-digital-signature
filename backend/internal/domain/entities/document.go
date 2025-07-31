package entities

import (
	"time"

	"github.com/google/uuid"
)

type Document struct {
	ID            string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID        string    `json:"user_id" gorm:"not null"`
	Filename      string    `json:"filename" gorm:"not null"`
	Issuer        string    `json:"issuer" gorm:"not null"`
	DocumentHash  string    `json:"document_hash" gorm:"not null"`
	SignatureData string    `json:"signature_data" gorm:"not null"`
	QRCodeData    string    `json:"qr_code_data" gorm:"not null"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	FileSize      int64     `json:"file_size"`
	Status        string    `json:"status" gorm:"default:active"`
	User          User      `json:"user" gorm:"foreignKey:UserID"`
}

type VerificationLog struct {
	ID                 string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	DocumentID         string    `json:"document_id"`
	VerificationResult string    `json:"verification_result"`
	VerifiedAt         time.Time `json:"verified_at"`
	VerifierIP         string    `json:"verifier_ip"`
	Details            string    `json:"details" gorm:"type:jsonb"`
	Document           Document  `json:"document" gorm:"foreignKey:DocumentID"`
}

func (d *Document) BeforeCreate() error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}

func (v *VerificationLog) BeforeCreate() error {
	if v.ID == "" {
		v.ID = uuid.New().String()
	}
	return nil
}