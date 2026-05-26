// Package subscription holds the subscription domain model (a customer's
// ownership of a catalog product) and the service that lists them. The service
// depends on the Repository interface, not on any database, so it can be
// unit-tested in isolation.
package subscription

import (
	"context"
	"time"

	"ispcrm/internal/product"
)

// Status is the lifecycle state of a subscription.
type Status string

const (
	StatusActive    Status = "active"
	StatusPending   Status = "pending"
	StatusCancelled Status = "cancelled"
)

// Subscription links a customer to a catalog product they own. The monthly
// price is snapshotted at signup so later catalog price changes don't alter
// what the customer is recorded as paying. The product name/category come from
// the live catalog via the preloaded Product association.
type Subscription struct {
	ID                   uint            `gorm:"primaryKey" json:"id"`
	CustomerID           uint            `json:"customerId"`
	ProductID            uint            `json:"productId"`
	Status               Status          `json:"status"`
	StartDate            time.Time       `json:"startDate"`
	EndDate              *time.Time      `json:"endDate,omitempty"`
	MonthlyPriceSnapshot float64         `json:"monthlyPriceSnapshot"`
	Quantity             int             `json:"quantity"`
	Product              product.Product `gorm:"foreignKey:ProductID" json:"product"`
}

// Repository is the persistence seam the service depends on.
type Repository interface {
	ListByCustomer(ctx context.Context, customerID uint) ([]Subscription, error)
}

// Service owns subscription business logic.
type Service struct {
	repo Repository
}

// NewService wires a Service to its repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ListForCustomer returns all subscriptions belonging to a customer.
func (s *Service) ListForCustomer(ctx context.Context, customerID uint) ([]Subscription, error) {
	return s.repo.ListByCustomer(ctx, customerID)
}
