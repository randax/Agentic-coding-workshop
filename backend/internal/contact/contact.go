// Package contact holds the contact domain model and service. A contact is a
// person associated with an account. The service depends on the Repository
// interface, not a database, so it is unit-testable in isolation.
package contact

import (
	"context"
	"errors"
	"strings"
)

// ErrNotFound is returned when a requested contact does not exist.
var ErrNotFound = errors.New("contact not found")

// ErrNameRequired is returned when a contact has no name.
var ErrNameRequired = errors.New("contact name is required")

// Contact is a person associated with an account. AssignedUserID/TeamID scope
// record visibility (own-or-team), matching the access model used elsewhere.
type Contact struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Title     string `json:"title"`
	AccountID uint   `gorm:"index" json:"accountId"`

	AssignedUserID *uint `json:"assignedUserId,omitempty"`
	TeamID         *uint `json:"teamId,omitempty"`
}

// Repository is the persistence seam the service depends on.
type Repository interface {
	List(ctx context.Context) ([]Contact, error)
	Get(ctx context.Context, id uint) (Contact, error)
	Create(ctx context.Context, c *Contact) error
	Update(ctx context.Context, c *Contact) error
	ListByAccount(ctx context.Context, accountID uint) ([]Contact, error)
}

// Service owns contact business logic.
type Service struct {
	repo Repository
}

// NewService wires a Service to its repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// List returns all contacts.
func (s *Service) List(ctx context.Context) ([]Contact, error) {
	return s.repo.List(ctx)
}

// ListByAccount returns the contacts belonging to one account.
func (s *Service) ListByAccount(ctx context.Context, accountID uint) ([]Contact, error) {
	return s.repo.ListByAccount(ctx, accountID)
}

// Get returns a single contact by ID, or ErrNotFound.
func (s *Service) Get(ctx context.Context, id uint) (Contact, error) {
	return s.repo.Get(ctx, id)
}

// Create adds a new contact, validating required fields.
func (s *Service) Create(ctx context.Context, c Contact) (Contact, error) {
	if strings.TrimSpace(c.Name) == "" {
		return Contact{}, ErrNameRequired
	}
	c.ID = 0
	if err := s.repo.Create(ctx, &c); err != nil {
		return Contact{}, err
	}
	return c, nil
}

// Update edits an existing contact's fields, preserving server-managed fields by
// loading the existing record first; an unknown ID yields ErrNotFound.
func (s *Service) Update(ctx context.Context, c Contact) (Contact, error) {
	existing, err := s.repo.Get(ctx, c.ID)
	if err != nil {
		return Contact{}, err
	}
	if strings.TrimSpace(c.Name) == "" {
		return Contact{}, ErrNameRequired
	}
	existing.Name = c.Name
	existing.Email = c.Email
	existing.Phone = c.Phone
	existing.Title = c.Title
	existing.AccountID = c.AccountID
	existing.AssignedUserID = c.AssignedUserID
	existing.TeamID = c.TeamID
	if err := s.repo.Update(ctx, &existing); err != nil {
		return Contact{}, err
	}
	return existing, nil
}
