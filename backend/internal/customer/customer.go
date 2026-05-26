// Package customer holds the customer domain model and the service that owns
// customer-related business logic. The service depends on the Repository
// interface, not on any database, so it can be unit-tested in isolation.
package customer

import (
	"context"
	"errors"
	"strings"
	"time"
)

// ErrNotFound is returned when a requested customer does not exist. Repository
// implementations translate storage-specific "not found" errors into this so
// the rest of the app stays decoupled from the persistence layer.
var ErrNotFound = errors.New("customer not found")

// Status is a customer's account standing.
type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
)

// Customer is a residential ISP customer. Struct tags configure GORM and JSON
// but the package itself does not import GORM, keeping the domain layer clean.
type Customer struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	ServiceAddress string    `json:"serviceAddress"`
	AccountNumber  string    `gorm:"uniqueIndex" json:"accountNumber"`
	CustomerSince  time.Time `json:"customerSince"`
	Status         Status    `json:"status"`
}

// Repository is the persistence seam the service depends on. The real
// implementation lives in the store package; tests use an in-memory fake.
type Repository interface {
	List(ctx context.Context) ([]Customer, error)
	// Get returns the customer with the given ID, or ErrNotFound.
	Get(ctx context.Context, id uint) (Customer, error)
}

// Filter narrows a customer list query. The zero value matches every customer.
type Filter struct {
	// Search is a partial, case-insensitive term matched against a customer's
	// name, email, or account number. Empty matches all.
	Search string
	// Status restricts results to a single account status. Empty matches all.
	Status Status
}

// Service owns customer business logic.
type Service struct {
	repo Repository
}

// NewService wires a Service to its repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// List returns the customers matching the filter. Filtering happens in the
// service (not the database) so the matching logic is unit-testable in isolation.
func (s *Service) List(ctx context.Context, f Filter) ([]Customer, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	var matched []Customer
	for _, c := range all {
		if f.matches(c) {
			matched = append(matched, c)
		}
	}
	return matched, nil
}

// matches reports whether a customer satisfies the filter.
func (f Filter) matches(c Customer) bool {
	if f.Status != "" && c.Status != f.Status {
		return false
	}
	if term := strings.ToLower(f.Search); term != "" {
		if !strings.Contains(strings.ToLower(c.Name), term) &&
			!strings.Contains(strings.ToLower(c.Email), term) &&
			!strings.Contains(strings.ToLower(c.AccountNumber), term) {
			return false
		}
	}
	return true
}

// Get returns a single customer by ID, or ErrNotFound.
func (s *Service) Get(ctx context.Context, id uint) (Customer, error) {
	return s.repo.Get(ctx, id)
}
