// Package customer holds the customer domain model and the service that owns
// customer-related business logic. The service depends on the Repository
// interface, not on any database, so it can be unit-tested in isolation.
package customer

import (
	"context"
	"time"
)

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
}

// Service owns customer business logic.
type Service struct {
	repo Repository
}

// NewService wires a Service to its repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// List returns all customers.
func (s *Service) List(ctx context.Context) ([]Customer, error) {
	return s.repo.List(ctx)
}
