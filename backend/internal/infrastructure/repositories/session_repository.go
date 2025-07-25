package repositories

import (
	"context"
	"digital-signature-system/internal/domain/entities"
	"digital-signature-system/internal/domain/repositories"
	"errors"

	"gorm.io/gorm"
)

type sessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *gorm.DB) repositories.SessionRepository {
	return &sessionRepository{
		db: db,
	}
}

func (r *sessionRepository) Create(ctx context.Context, session *entities.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *sessionRepository) GetByToken(ctx context.Context, token string) (*entities.Session, error) {
	var session entities.Session
	if err := r.db.WithContext(ctx).Where("session_token = ?", token).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) Update(ctx context.Context, session *entities.Session) error {
	return r.db.WithContext(ctx).Save(session).Error
}

func (r *sessionRepository) Delete(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Delete(&entities.Session{}, "session_token = ?", token).Error
}

func (r *sessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Delete(&entities.Session{}, "user_id = ?", userID).Error
}