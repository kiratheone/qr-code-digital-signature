package repositories

import (
	"context"

	"digital-signature-system/internal/domain/entities"
)

type SessionRepository interface {
	Create(ctx context.Context, session *entities.Session) error
	GetByToken(ctx context.Context, token string) (*entities.Session, error)
	GetByUserID(ctx context.Context, userID string) ([]*entities.Session, error)
	Update(ctx context.Context, session *entities.Session) error
	Delete(ctx context.Context, token string) error
	DeleteByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
}