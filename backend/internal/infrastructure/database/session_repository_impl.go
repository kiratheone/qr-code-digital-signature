package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"digital-signature-system/internal/domain/entities"
	"digital-signature-system/internal/domain/repositories"
)

type sessionRepositoryImpl struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) repositories.SessionRepository {
	return &sessionRepositoryImpl{db: db}
}

func (r *sessionRepositoryImpl) Create(ctx context.Context, session *entities.Session) error {
	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

func (r *sessionRepositoryImpl) GetByToken(ctx context.Context, token string) (*entities.Session, error) {
	var session entities.Session
	if err := r.db.WithContext(ctx).Preload("User").Where("session_token = ?", token).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session by token: %w", err)
	}
	return &session, nil
}

func (r *sessionRepositoryImpl) GetByUserID(ctx context.Context, userID string) ([]*entities.Session, error) {
	var sessions []*entities.Session
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("failed to get sessions by user ID: %w", err)
	}
	return sessions, nil
}

func (r *sessionRepositoryImpl) Update(ctx context.Context, session *entities.Session) error {
	if err := r.db.WithContext(ctx).Save(session).Error; err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}
	return nil
}

func (r *sessionRepositoryImpl) Delete(ctx context.Context, token string) error {
	if err := r.db.WithContext(ctx).Delete(&entities.Session{}, "session_token = ?", token).Error; err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (r *sessionRepositoryImpl) DeleteByUserID(ctx context.Context, userID string) error {
	if err := r.db.WithContext(ctx).Delete(&entities.Session{}, "user_id = ?", userID).Error; err != nil {
		return fmt.Errorf("failed to delete sessions by user ID: %w", err)
	}
	return nil
}

func (r *sessionRepositoryImpl) DeleteExpired(ctx context.Context) error {
	if err := r.db.WithContext(ctx).Delete(&entities.Session{}, "expires_at < ?", time.Now()).Error; err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}
	return nil
}