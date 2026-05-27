package store

import (
	"context"
	"time"

	"saltcrm/internal/identity"

	"gorm.io/gorm"
)

// SessionRepository is the GORM-backed implementation of
// identity.SessionRepository.
type SessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository wires a repository to a GORM database handle.
func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create persists a new session.
func (r *SessionRepository) Create(ctx context.Context, s *identity.Session) error {
	return r.db.WithContext(ctx).Create(s).Error
}

// FindUserID returns the user id for an unexpired session token, or
// identity.ErrNoSession if the token is unknown or expired.
func (r *SessionRepository) FindUserID(ctx context.Context, token string) (uint, error) {
	var s identity.Session
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&s).Error
	if err != nil || s.ExpiresAt.Before(time.Now()) {
		return 0, identity.ErrNoSession
	}
	return s.UserID, nil
}

// Delete removes a session token (logout); deleting an unknown token is a no-op.
func (r *SessionRepository) Delete(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token = ?", token).Delete(&identity.Session{}).Error
}
