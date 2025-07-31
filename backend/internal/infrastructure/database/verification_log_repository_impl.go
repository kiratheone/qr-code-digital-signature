package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"digital-signature-system/internal/domain/entities"
	"digital-signature-system/internal/domain/repositories"
)

type verificationLogRepositoryImpl struct {
	db *gorm.DB
}

func NewVerificationLogRepository(db *gorm.DB) repositories.VerificationLogRepository {
	return &verificationLogRepositoryImpl{db: db}
}

func (r *verificationLogRepositoryImpl) Create(ctx context.Context, log *entities.VerificationLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return fmt.Errorf("failed to create verification log: %w", err)
	}
	return nil
}

func (r *verificationLogRepositoryImpl) GetByDocumentID(ctx context.Context, docID string) ([]*entities.VerificationLog, error) {
	var logs []*entities.VerificationLog
	if err := r.db.WithContext(ctx).Preload("Document").Where("document_id = ?", docID).Order("verified_at DESC").Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to get verification logs by document ID: %w", err)
	}
	return logs, nil
}

func (r *verificationLogRepositoryImpl) GetByID(ctx context.Context, id string) (*entities.VerificationLog, error) {
	var log entities.VerificationLog
	if err := r.db.WithContext(ctx).Preload("Document").Where("id = ?", id).First(&log).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get verification log by ID: %w", err)
	}
	return &log, nil
}