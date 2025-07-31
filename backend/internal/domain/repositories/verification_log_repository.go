package repositories

import (
	"context"

	"digital-signature-system/internal/domain/entities"
)

type VerificationLogRepository interface {
	Create(ctx context.Context, log *entities.VerificationLog) error
	GetByDocumentID(ctx context.Context, docID string) ([]*entities.VerificationLog, error)
	GetByID(ctx context.Context, id string) (*entities.VerificationLog, error)
}