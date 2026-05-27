// Package agent holds the support-agent domain model and a service that lists
// them. Agents are seeded (there is no authentication yet); they exist so cases
// can be assigned and comments attributed to a named person. The service
// depends on the Repository interface, not on any database, so it can be
// unit-tested in isolation.
package agent

import (
	"context"
	"errors"
)

// ErrNotFound is returned when a requested agent does not exist.
var ErrNotFound = errors.New("agent not found")

// Role is an agent's access role, gating which actions they may perform.
type Role string

const (
	RoleAdmin   Role = "admin"
	RoleManager Role = "manager"
	RoleAgent   Role = "agent"
)

// Valid reports whether r is a known role.
func (r Role) Valid() bool {
	switch r {
	case RoleAdmin, RoleManager, RoleAgent:
		return true
	default:
		return false
	}
}

// Agent is a customer-service agent and the application's login user. It is
// referenced by cases as assignee and by comments as author. PasswordHash is
// never serialized. TeamID scopes record visibility (enforced in a later slice).
type Agent struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Name         string `json:"name"`
	Email        string `gorm:"uniqueIndex" json:"email"`
	PasswordHash string `json:"-"`
	Role         Role   `json:"role"`
	TeamID       *uint  `json:"teamId,omitempty"`
}

// Repository is the persistence seam the service depends on.
type Repository interface {
	List(ctx context.Context) ([]Agent, error)
}

// Service owns agent business logic.
type Service struct {
	repo Repository
}

// NewService wires a Service to its repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// List returns all agents.
func (s *Service) List(ctx context.Context) ([]Agent, error) {
	return s.repo.List(ctx)
}
