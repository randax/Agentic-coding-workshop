// Package conversion owns the lead→account conversion workflow: the cross-domain
// orchestration that promotes a qualified lead into a new account with a linked
// contact. It holds all the business rules (eligibility, visibility, field
// mapping) behind a small Repository interface it defines itself.
//
// Per the project's dependency rule it imports only pure domain models
// (lead, customer, contact, agent) and the access package — never GORM, Gin,
// internal/store, or the other domain services. The store implements Persist
// inside a single transaction so the whole conversion is all-or-nothing.
package conversion

import (
	"context"
	"errors"

	"saltcrm/internal/access"
	"saltcrm/internal/agent"
	"saltcrm/internal/contact"
	"saltcrm/internal/customer"
	"saltcrm/internal/lead"
)

// ErrNotFound is returned when the lead does not exist or is not visible to the
// viewer. Not-visible is reported as not-found so conversion can't reveal the
// existence of records outside the viewer's scope. Maps to HTTP 404.
var ErrNotFound = errors.New("lead not found")

// ErrNotQualified is returned when the lead is not in the qualified status, the
// only status from which conversion is allowed. Maps to HTTP 409.
var ErrNotQualified = errors.New("lead is not qualified")

// ErrAlreadyConverted is returned when the lead has already been converted
// (its ConvertedAccountID is set). Maps to HTTP 409.
var ErrAlreadyConverted = errors.New("lead already converted")

// Plan is the fully field-mapped set of records a single conversion creates. The
// service builds it; the repository persists it atomically. Account.AccountNumber
// and the foreign keys (Account.ID, Contact.AccountID) are filled inside the
// persisting transaction, where uniqueness and ordering can be guaranteed.
type Plan struct {
	LeadID  uint
	Account customer.Customer
	Contact contact.Contact
}

// Result is the outcome of a conversion: the ids of the records created.
type Result struct {
	AccountID uint
	ContactID uint
}

// Repository is the persistence seam the service depends on. The real
// implementation lives in the store package and runs Persist in one transaction;
// tests use an in-memory fake.
type Repository interface {
	// GetLead returns the lead to convert, or lead.ErrNotFound.
	GetLead(ctx context.Context, id uint) (lead.Lead, error)
	// Persist creates the account + contact and flips the lead, all atomically,
	// returning the new record ids.
	Persist(ctx context.Context, plan Plan) (Result, error)
}

// Service owns the conversion business rules.
type Service struct {
	repo Repository
}

// NewService wires a Service to its repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Convert promotes a qualified lead, visible to the viewer, into a new account
// with a linked contact. It enforces visibility (own-or-team), eligibility (only
// qualified, never twice), maps the lead's fields onto the new records, and hands
// the plan to the repository to persist atomically.
func (s *Service) Convert(ctx context.Context, viewer agent.Agent, leadID uint) (Result, error) {
	l, err := s.repo.GetLead(ctx, leadID)
	if err != nil {
		if errors.Is(err, lead.ErrNotFound) {
			return Result{}, ErrNotFound
		}
		return Result{}, err
	}
	if !access.Visible(viewer, l.AssignedUserID, l.TeamID) {
		return Result{}, ErrNotFound
	}
	if l.ConvertedAccountID != nil {
		return Result{}, ErrAlreadyConverted
	}
	if l.Status != lead.StatusQualified {
		return Result{}, ErrNotQualified
	}

	plan := Plan{
		LeadID: l.ID,
		Account: customer.Customer{
			Name:           l.Company,
			Email:          l.Email,
			Phone:          l.Phone,
			Status:         customer.StatusActive,
			AssignedUserID: l.AssignedUserID,
			TeamID:         l.TeamID,
		},
		Contact: contact.Contact{
			Name:           l.Name,
			Email:          l.Email,
			Phone:          l.Phone,
			AssignedUserID: l.AssignedUserID,
			TeamID:         l.TeamID,
		},
	}
	return s.repo.Persist(ctx, plan)
}
