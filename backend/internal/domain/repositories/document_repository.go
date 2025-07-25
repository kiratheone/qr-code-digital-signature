package repositories

import (
	"context"
	"digital-signature-system/internal/domain/entities"
)

type DocumentFilter struct {
	Search   string
	Limit    int
	Offset   int
	SortBy   string
	SortDesc bool
}

type DocumentRepository interface {
	Create(ctx context.Context, doc *entities.Document) error
	GetByID(ctx context.Context, id string) (*entities.Document, error)
	GetByUserID(ctx context.Context, userID string, filter DocumentFilter) ([]*entities.Document, int64, error)
	Update(ctx context.Context, doc *entities.Document) error
	Delete(ctx context.Context, id string) error
}