// Package identity owns authentication: verifying credentials, hashing
// passwords, and issuing/resolving login sessions. The login user is an
// agent.Agent. The service depends on repository interfaces, not a database,
// so it can be unit-tested in isolation.
package identity

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"saltcrm/internal/agent"

	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidCredentials is returned when an email/password pair does not match.
var ErrInvalidCredentials = errors.New("invalid credentials")

// ErrNoSession is returned when a session token is unknown or expired.
var ErrNoSession = errors.New("no active session")

// sessionTTL is how long a login session stays valid.
const sessionTTL = 7 * 24 * time.Hour

// Session is a login session: an opaque token bound to a user, with an expiry.
type Session struct {
	Token     string    `gorm:"primaryKey"`
	UserID    uint      `gorm:"index"`
	ExpiresAt time.Time `gorm:"index"`
}

// UserRepository looks up login users (agents) by email or id.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (agent.Agent, error)
	FindByID(ctx context.Context, id uint) (agent.Agent, error)
}

// SessionRepository persists login sessions. FindUserID must return ErrNoSession
// for tokens that are unknown or expired.
type SessionRepository interface {
	Create(ctx context.Context, s *Session) error
	FindUserID(ctx context.Context, token string) (uint, error)
	Delete(ctx context.Context, token string) error
}

// Service owns authentication logic.
type Service struct {
	users    UserRepository
	sessions SessionRepository
}

// NewService wires a Service to its repositories.
func NewService(users UserRepository, sessions SessionRepository) *Service {
	return &Service{users: users, sessions: sessions}
}

// HashPassword returns a bcrypt hash suitable for storing on a user. Used by
// seeding and any future user-creation flow.
func HashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(h), err
}

// Authenticate verifies an email/password pair, returning the user or
// ErrInvalidCredentials. Unknown emails and wrong passwords are indistinguishable.
func (s *Service) Authenticate(ctx context.Context, email, password string) (agent.Agent, error) {
	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return agent.Agent{}, ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return agent.Agent{}, ErrInvalidCredentials
	}
	return u, nil
}

// Login authenticates the credentials and, on success, creates a session,
// returning the user and the session token.
func (s *Service) Login(ctx context.Context, email, password string) (agent.Agent, string, error) {
	u, err := s.Authenticate(ctx, email, password)
	if err != nil {
		return agent.Agent{}, "", err
	}
	token, err := newToken()
	if err != nil {
		return agent.Agent{}, "", err
	}
	session := &Session{Token: token, UserID: u.ID, ExpiresAt: time.Now().Add(sessionTTL)}
	if err := s.sessions.Create(ctx, session); err != nil {
		return agent.Agent{}, "", err
	}
	return u, token, nil
}

// CurrentUser resolves a session token to its user, or ErrNoSession.
func (s *Service) CurrentUser(ctx context.Context, token string) (agent.Agent, error) {
	userID, err := s.sessions.FindUserID(ctx, token)
	if err != nil {
		return agent.Agent{}, err
	}
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return agent.Agent{}, ErrNoSession
	}
	return u, nil
}

// Logout invalidates a session token.
func (s *Service) Logout(ctx context.Context, token string) error {
	return s.sessions.Delete(ctx, token)
}

// newToken returns a cryptographically-random opaque session token.
func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
