package repositories

import (
	"context"
	"digital-signature-system/internal/domain/entities"
	"digital-signature-system/internal/domain/repositories"
	"errors"

	"gorm.io/gorm"
)

type verificationLogRepository struct {
	db *gorm.DB
}

// NewVerificationLogRepository creates a new verification log repository
func NewVerificationLogRepository(db *gorm.DB) repositories.VerificationLogRepository {
	return &verificationLogRepository{
		db: db,
	}
}

func (r *verificationLogRepository) Create(ctx context.Context, log *entities.VerificationLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *verificationLogRepository) GetByDocumentID(ctx context.Context, docID string) ([]*entities.VerificationLog, error) {
	var logs []*entities.VerificationLog
	if err := r.db.WithContext(ctx).Where("document_id = ?", docID).Order("verified_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *verificationLogRepository) GetByID(ctx context.Context, id string) (*entities.VerificationLog, error) {
	var log entities.VerificationLog
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&log).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &log, nil
}