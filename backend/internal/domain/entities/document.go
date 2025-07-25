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

func NewDocument(userID, filename, issuer, documentHash, signatureData, qrCodeData string, fileSize int64) *Document {
	return &Document{
		ID:            uuid.New().String(),
		UserID:        userID,
		Filename:      filename,
		Issuer:        issuer,
		DocumentHash:  documentHash,
		SignatureData: signatureData,
		QRCodeData:    qrCodeData,
		FileSize:      fileSize,
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}