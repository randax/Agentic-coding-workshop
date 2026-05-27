package report

import (
	"context"
	"errors"
	"strings"
)

// Validation and lookup errors for saved reports.
var (
	ErrNameRequired   = errors.New("report name is required")
	ErrModuleRequired = errors.New("report module is required")
	ErrNotFound       = errors.New("report not found")
)

// Saved is a persisted report: a named, re-runnable Definition. The definition
// is stored as a JSON column so the whole query (filters, grouping, aggregation)
// round-trips without a bespoke schema.
type Saved struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	Name       string     `json:"name"`
	Definition Definition `gorm:"serializer:json" json:"definition"`
}

// Repository is the persistence seam the service depends on.
type Repository interface {
	Create(ctx context.Context, s *Saved) error
	List(ctx context.Context) ([]Saved, error)
	Get(ctx context.Context, id uint) (Saved, error)
}

// Service owns saved-report persistence. The aggregation itself stays in the
// pure Run function; this service only stores and retrieves definitions.
type Service struct {
	repo Repository
}

// NewService wires a Service to its repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Save validates and persists a report, returning it with its assigned ID.
func (s *Service) Save(ctx context.Context, r Saved) (Saved, error) {
	if strings.TrimSpace(r.Name) == "" {
		return Saved{}, ErrNameRequired
	}
	if strings.TrimSpace(r.Definition.Module) == "" {
		return Saved{}, ErrModuleRequired
	}
	r.ID = 0
	if err := s.repo.Create(ctx, &r); err != nil {
		return Saved{}, err
	}
	return r, nil
}

// List returns every saved report.
func (s *Service) List(ctx context.Context) ([]Saved, error) {
	return s.repo.List(ctx)
}

// Get returns a saved report by ID, or ErrNotFound.
func (s *Service) Get(ctx context.Context, id uint) (Saved, error) {
	return s.repo.Get(ctx, id)
}
