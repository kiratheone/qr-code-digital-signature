package entities

import (
	"time"

	"github.com/google/uuid"
)

type VerificationLog struct {
	ID                 string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	DocumentID         string    `json:"document_id"`
	VerificationResult string    `json:"verification_result"`
	VerifiedAt         time.Time `json:"verified_at"`
	VerifierIP         string    `json:"verifier_ip"`
	Details            string    `json:"details" gorm:"type:jsonb"`
	Document           Document  `json:"document" gorm:"foreignKey:DocumentID"`
}

func NewVerificationLog(documentID, result, verifierIP, details string) *VerificationLog {
	return &VerificationLog{
		ID:                 uuid.New().String(),
		DocumentID:         documentID,
		VerificationResult: result,
		VerifierIP:         verifierIP,
		Details:            details,
		VerifiedAt:         time.Now(),
	}
}