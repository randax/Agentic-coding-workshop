// Package agent holds the support-agent domain model and a service that lists
// them. Agents are seeded (there is no authentication yet); they exist so cases
// can be assigned and comments attributed to a named person. The service
// depends on the Repository interface, not on any database, so it can be
// unit-tested in isolation.
package agent

import "context"

// Agent is a customer-service agent. There is no login; agents are a seeded
// reference list used as case assignees and comment authors.
type Agent struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
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
