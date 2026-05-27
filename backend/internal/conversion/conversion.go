// Package conversion turns a qualified lead into an account, a contact, and an
// opportunity in one atomic step. The mapping logic lives in the service; the
// Repository persists the plan transactionally so a partial failure rolls back.
package conversion

import (
	"context"
	"errors"
	"strings"

	"saltcrm/internal/contact"
	"saltcrm/internal/customer"
	"saltcrm/internal/lead"
	"saltcrm/internal/opportunity"
)

// ErrAlreadyConverted is returned when converting a lead that is already converted.
var ErrAlreadyConverted = errors.New("lead is already converted")

// Result is the set of records a conversion created.
type Result struct {
	AccountID     uint `json:"accountId"`
	ContactID     uint `json:"contactId"`
	OpportunityID uint `json:"opportunityId"`
}

// Plan is the set of records to create from a lead, handed to the repository to
// persist atomically. The repository fills in cross-references (the contact and
// opportunity get the new account's ID) and marks the lead converted.
type Plan struct {
	Lead        lead.Lead
	Account     customer.Customer
	Contact     contact.Contact
	Opportunity opportunity.Opportunity
}

// Repository persists a conversion plan atomically and loads leads.
type Repository interface {
	GetLead(ctx context.Context, id uint) (lead.Lead, error)
	// Persist creates the account, then the contact + opportunity referencing it,
	// marks the lead converted (linking the created records), all in one
	// transaction, returning the new IDs.
	Persist(ctx context.Context, plan Plan) (Result, error)
}

// Service owns lead-conversion logic.
type Service struct {
	repo Repository
}

// NewService wires a Service to its repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Convert turns the given lead into an account + contact + opportunity. It is
// rejected if the lead is already converted, and yields lead.ErrNotFound for an
// unknown lead.
func (s *Service) Convert(ctx context.Context, leadID uint) (Result, error) {
	l, err := s.repo.GetLead(ctx, leadID)
	if err != nil {
		return Result{}, err
	}
	if l.Status == lead.StatusConverted {
		return Result{}, ErrAlreadyConverted
	}

	accountName := strings.TrimSpace(l.Company)
	if accountName == "" {
		accountName = l.Name // a lead without a company becomes a personal account
	}

	plan := Plan{
		Lead: l,
		Account: customer.Customer{
			Name:           accountName,
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
		Opportunity: opportunity.Opportunity{
			Name:           accountName + " — new opportunity",
			Stage:          opportunity.StageProspecting,
			Probability:    opportunity.StageProspecting.Probability(),
			AssignedUserID: l.AssignedUserID,
			TeamID:         l.TeamID,
		},
	}
	return s.repo.Persist(ctx, plan)
}
