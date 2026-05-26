// Package supportcase holds the support-case domain model and the service that
// owns case business logic. (The package is named supportcase because "case" is
// a Go keyword.) This slice is view-only — listing a customer's cases; the case
// lifecycle state machine and mutations arrive in later slices. The service
// depends on the Repository interface, not on any database, so it can be
// unit-tested in isolation.
package supportcase

import (
	"context"
	"errors"
	"time"

	"ispcrm/internal/agent"
)

// ErrNotFound is returned when a requested case does not exist.
var ErrNotFound = errors.New("case not found")

// Category is the kind of issue a case is about.
type Category string

const (
	CategoryBilling      Category = "billing"
	CategoryConnectivity Category = "connectivity"
	CategoryHardware     Category = "hardware"
	CategoryTV           Category = "tv"
	CategoryGeneral      Category = "general"
)

// Priority is how urgently a case needs attention.
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

// Status is a case's position in its lifecycle (Open → In-progress → Resolved → Closed).
type Status string

const (
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusResolved   Status = "resolved"
	StatusClosed     Status = "closed"
)

// Case is a support case/complaint a customer has filed. The assigned agent is a
// nullable reference, preloaded for display. CreatedAt/UpdatedAt are managed by GORM.
type Case struct {
	ID              uint          `gorm:"primaryKey" json:"id"`
	CustomerID      uint          `json:"customerId"`
	Subject         string        `json:"subject"`
	Description     string        `json:"description"`
	Category        Category      `json:"category"`
	Priority        Priority      `json:"priority"`
	Status          Status        `json:"status"`
	AssignedAgentID *uint         `json:"assignedAgentId,omitempty"`
	AssignedAgent   *agent.Agent  `gorm:"foreignKey:AssignedAgentID" json:"assignedAgent,omitempty"`
	Comments        []CaseComment `gorm:"foreignKey:CaseID" json:"comments,omitempty"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
}

// CaseComment is a single entry on a case's chronological timeline: a body, the
// agent who authored it, and a timestamp. The author is preloaded for display.
type CaseComment struct {
	ID            uint         `gorm:"primaryKey" json:"id"`
	CaseID        uint         `json:"caseId"`
	Body          string       `json:"body"`
	AuthorAgentID *uint        `json:"authorAgentId,omitempty"`
	AuthorAgent   *agent.Agent `gorm:"foreignKey:AuthorAgentID" json:"authorAgent,omitempty"`
	CreatedAt     time.Time    `json:"createdAt"`
}

// Repository is the persistence seam the service depends on.
type Repository interface {
	ListByCustomer(ctx context.Context, customerID uint) ([]Case, error)
	// Get returns a case with its comment timeline, or ErrNotFound.
	Get(ctx context.Context, id uint) (Case, error)
}

// Service owns support-case business logic.
type Service struct {
	repo Repository
}

// NewService wires a Service to its repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ListForCustomer returns all cases filed by a customer.
func (s *Service) ListForCustomer(ctx context.Context, customerID uint) ([]Case, error) {
	return s.repo.ListByCustomer(ctx, customerID)
}

// Get returns a single case with its comment timeline, or ErrNotFound.
func (s *Service) Get(ctx context.Context, id uint) (Case, error) {
	return s.repo.Get(ctx, id)
}
