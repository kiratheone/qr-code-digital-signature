package repositories

import (
	"context"
	"digital-signature-system/internal/domain/entities"
	"digital-signature-system/internal/domain/repositories"
	"errors"

	"gorm.io/gorm"
)

type documentRepository struct {
	db *gorm.DB
}

// NewDocumentRepository creates a new document repository
func NewDocumentRepository(db *gorm.DB) repositories.DocumentRepository {
	return &documentRepository{
		db: db,
	}
}

func (r *documentRepository) Create(ctx context.Context, doc *entities.Document) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

func (r *documentRepository) GetByID(ctx context.Context, id string) (*entities.Document, error) {
	var doc entities.Document
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&doc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &doc, nil
}

func (r *documentRepository) GetByUserID(ctx context.Context, userID string, filter repositories.DocumentFilter) ([]*entities.Document, int64, error) {
	var docs []*entities.Document
	var total int64

	query := r.db.WithContext(ctx).Model(&entities.Document{}).Where("user_id = ?", userID)

	// Apply search filter
	if filter.Search != "" {
		query = query.Where("filename ILIKE ? OR issuer ILIKE ?", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	// Count total before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	if filter.SortBy != "" {
		direction := "ASC"
		if filter.SortDesc {
			direction = "DESC"
		}
		query = query.Order(filter.SortBy + " " + direction)
	} else {
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	if err := query.Find(&docs).Error; err != nil {
		return nil, 0, err
	}

	return docs, total, nil
}

func (r *documentRepository) Update(ctx context.Context, doc *entities.Document) error {
	return r.db.WithContext(ctx).Save(doc).Error
}

func (r *documentRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entities.Document{}, "id = ?", id).Error
}