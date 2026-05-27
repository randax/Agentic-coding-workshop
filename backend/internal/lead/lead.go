// Package lead holds the lead domain model and service. A lead is an
// unqualified prospect that can later be converted into an account + contact +
// opportunity. The service depends on the Repository interface, not a database.
package lead

import (
	"context"
	"errors"
	"strings"
)

// ErrNotFound is returned when a requested lead does not exist.
var ErrNotFound = errors.New("lead not found")

// ErrNameRequired is returned when a lead has no name.
var ErrNameRequired = errors.New("lead name is required")

// ErrInvalidStatus is returned when a lead's status is not a known value.
var ErrInvalidStatus = errors.New("invalid lead status")

// Status is a lead's position in the qualification funnel.
type Status string

const (
	StatusNew         Status = "new"
	StatusWorking     Status = "working"
	StatusQualified   Status = "qualified"
	StatusUnqualified Status = "unqualified"
	StatusConverted   Status = "converted"
)

// Valid reports whether s is a known lead status.
func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusWorking, StatusQualified, StatusUnqualified, StatusConverted:
		return true
	default:
		return false
	}
}

// Lead is an unqualified prospect. AssignedUserID/TeamID scope visibility.
type Lead struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Company string `json:"company"`
	Status  Status `json:"status"`

	AssignedUserID *uint `json:"assignedUserId,omitempty"`
	TeamID         *uint `json:"teamId,omitempty"`
}

// Repository is the persistence seam the service depends on.
type Repository interface {
	List(ctx context.Context) ([]Lead, error)
	Get(ctx context.Context, id uint) (Lead, error)
	Create(ctx context.Context, l *Lead) error
	Update(ctx context.Context, l *Lead) error
}

// Service owns lead business logic.
type Service struct {
	repo Repository
}

// NewService wires a Service to its repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// List returns all leads.
func (s *Service) List(ctx context.Context) ([]Lead, error) {
	return s.repo.List(ctx)
}

// Get returns a single lead by ID, or ErrNotFound.
func (s *Service) Get(ctx context.Context, id uint) (Lead, error) {
	return s.repo.Get(ctx, id)
}

// Create adds a new lead. A new lead defaults to status "new".
func (s *Service) Create(ctx context.Context, l Lead) (Lead, error) {
	if l.Status == "" {
		l.Status = StatusNew
	}
	if strings.TrimSpace(l.Name) == "" {
		return Lead{}, ErrNameRequired
	}
	if !l.Status.Valid() {
		return Lead{}, ErrInvalidStatus
	}
	l.ID = 0
	if err := s.repo.Create(ctx, &l); err != nil {
		return Lead{}, err
	}
	return l, nil
}

// Update edits an existing lead, loading the existing record first so an
// unknown ID yields ErrNotFound.
func (s *Service) Update(ctx context.Context, l Lead) (Lead, error) {
	existing, err := s.repo.Get(ctx, l.ID)
	if err != nil {
		return Lead{}, err
	}
	if strings.TrimSpace(l.Name) == "" {
		return Lead{}, ErrNameRequired
	}
	if !l.Status.Valid() {
		return Lead{}, ErrInvalidStatus
	}
	existing.Name = l.Name
	existing.Email = l.Email
	existing.Phone = l.Phone
	existing.Company = l.Company
	existing.Status = l.Status
	existing.AssignedUserID = l.AssignedUserID
	existing.TeamID = l.TeamID
	if err := s.repo.Update(ctx, &existing); err != nil {
		return Lead{}, err
	}
	return existing, nil
}
