// Package subscription holds the subscription domain model (a customer's
// ownership of a catalog product) and the service that lists them. The service
// depends on the Repository interface, not on any database, so it can be
// unit-tested in isolation.
package subscription

import (
	"context"
	"errors"
	"time"

	"saltcrm/internal/product"
)

// Status is the lifecycle state of a subscription.
type Status string

const (
	StatusActive    Status = "active"
	StatusPending   Status = "pending"
	StatusCancelled Status = "cancelled"
)

// ErrNotFound is returned when a requested subscription does not exist.
var ErrNotFound = errors.New("subscription not found")

// ErrProductRetired is returned when assigning a product that is no longer
// available in the catalog.
var ErrProductRetired = errors.New("cannot subscribe to a retired product")

// ErrInvalidQuantity is returned when an assignment quantity is less than one.
var ErrInvalidQuantity = errors.New("quantity must be at least 1")

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
	List(ctx context.Context) ([]Subscription, error)
	ListByCustomer(ctx context.Context, customerID uint) ([]Subscription, error)
	Get(ctx context.Context, id uint) (Subscription, error)
	Create(ctx context.Context, s *Subscription) error
	Update(ctx context.Context, s *Subscription) error
}

// ProductReader is the slice of the catalog the service needs: it reads a
// product to snapshot its price and check it is still available.
type ProductReader interface {
	Get(ctx context.Context, id uint) (product.Product, error)
}

// Service owns subscription business logic.
type Service struct {
	repo     Repository
	products ProductReader
}

// NewService wires a Service to its repository and the catalog it reads from.
func NewService(repo Repository, products ProductReader) *Service {
	return &Service{repo: repo, products: products}
}

// Assign subscribes a customer to a catalog product, snapshotting the product's
// current monthly price so later catalog changes don't alter what the customer
// is recorded as paying. Retired products cannot be subscribed to.
func (s *Service) Assign(ctx context.Context, customerID, productID uint, quantity int) (Subscription, error) {
	if quantity < 1 {
		return Subscription{}, ErrInvalidQuantity
	}
	p, err := s.products.Get(ctx, productID)
	if err != nil {
		return Subscription{}, err
	}
	if !p.Available {
		return Subscription{}, ErrProductRetired
	}
	sub := Subscription{
		CustomerID:           customerID,
		ProductID:            productID,
		Status:               StatusActive,
		StartDate:            time.Now(),
		MonthlyPriceSnapshot: p.MonthlyPrice,
		Quantity:             quantity,
	}
	if err := s.repo.Create(ctx, &sub); err != nil {
		return Subscription{}, err
	}
	return sub, nil
}

// Cancel ends a subscription: it sets the status to cancelled and records the
// end date. Cancelling an already-cancelled subscription is a no-op that leaves
// the existing end date untouched.
func (s *Service) Cancel(ctx context.Context, id uint) (Subscription, error) {
	sub, err := s.repo.Get(ctx, id)
	if err != nil {
		return Subscription{}, err
	}
	if sub.Status == StatusCancelled {
		return sub, nil
	}
	now := time.Now()
	sub.Status = StatusCancelled
	sub.EndDate = &now
	if err := s.repo.Update(ctx, &sub); err != nil {
		return Subscription{}, err
	}
	return sub, nil
}

// ListForCustomer returns all subscriptions belonging to a customer.
func (s *Service) ListForCustomer(ctx context.Context, customerID uint) ([]Subscription, error) {
	return s.repo.ListByCustomer(ctx, customerID)
}

// List returns all subscriptions across customers.
func (s *Service) List(ctx context.Context) ([]Subscription, error) {
	return s.repo.List(ctx)
}
