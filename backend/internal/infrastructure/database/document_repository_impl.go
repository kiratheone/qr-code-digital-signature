package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"digital-signature-system/internal/domain/entities"
	"digital-signature-system/internal/domain/repositories"
)

type documentRepositoryImpl struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) repositories.DocumentRepository {
	return &documentRepositoryImpl{db: db}
}

func (r *documentRepositoryImpl) Create(ctx context.Context, doc *entities.Document) error {
	if err := r.db.WithContext(ctx).Create(doc).Error; err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}
	return nil
}

func (r *documentRepositoryImpl) GetByID(ctx context.Context, id string) (*entities.Document, error) {
	var doc entities.Document
	if err := r.db.WithContext(ctx).Preload("User").Where("id = ?", id).First(&doc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get document by ID: %w", err)
	}
	return &doc, nil
}

func (r *documentRepositoryImpl) GetByUserID(ctx context.Context, userID string, filter repositories.DocumentFilter) ([]*entities.Document, int64, error) {
	var docs []*entities.Document
	var total int64

	query := r.db.WithContext(ctx).Where("user_id = ?", userID)

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	// Count total records
	if err := query.Model(&entities.Document{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	// Apply pagination
	if filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	// Get documents with user preloaded
	if err := query.Preload("User").Order("created_at DESC").Find(&docs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get documents by user ID: %w", err)
	}

	return docs, total, nil
}

func (r *documentRepositoryImpl) GetByHash(ctx context.Context, hash string) (*entities.Document, error) {
	var doc entities.Document
	if err := r.db.WithContext(ctx).Preload("User").Where("document_hash = ?", hash).First(&doc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get document by hash: %w", err)
	}
	return &doc, nil
}

func (r *documentRepositoryImpl) Update(ctx context.Context, doc *entities.Document) error {
	if err := r.db.WithContext(ctx).Save(doc).Error; err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}
	return nil
}

func (r *documentRepositoryImpl) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&entities.Document{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	return nil
}