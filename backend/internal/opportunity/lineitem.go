package opportunity

import (
	"context"
	"errors"
)

// ErrInvalidQuantity is returned when a line item's quantity is not positive.
var ErrInvalidQuantity = errors.New("line item quantity must be positive")

// LineItem is a product on an opportunity, with the product name and unit price
// snapshotted at the time it was added so later catalog changes don't alter the
// recorded deal. LineTotal is the rolled-up UnitPrice*Quantity.
type LineItem struct {
	ID            uint    `gorm:"primaryKey" json:"id"`
	OpportunityID uint    `gorm:"index" json:"opportunityId"`
	ProductID     uint    `json:"productId"`
	ProductName   string  `json:"productName"`
	UnitPrice     float64 `json:"unitPrice"`
	Quantity      int     `json:"quantity"`
	LineTotal     float64 `json:"lineTotal"`
}

// LineItemRepository is the persistence seam for line items.
type LineItemRepository interface {
	Create(ctx context.Context, li *LineItem) error
	ListByOpportunity(ctx context.Context, opportunityID uint) ([]LineItem, error)
}

// LineItemService owns line-item business logic.
type LineItemService struct {
	repo LineItemRepository
}

// NewLineItemService wires a service to its repository.
func NewLineItemService(repo LineItemRepository) *LineItemService {
	return &LineItemService{repo: repo}
}

// Add appends a line item to an opportunity, computing its rolled-up total.
func (s *LineItemService) Add(ctx context.Context, li LineItem) (LineItem, error) {
	if li.Quantity <= 0 {
		return LineItem{}, ErrInvalidQuantity
	}
	li.ID = 0
	li.LineTotal = li.UnitPrice * float64(li.Quantity)
	if err := s.repo.Create(ctx, &li); err != nil {
		return LineItem{}, err
	}
	return li, nil
}

// ListByOpportunity returns the line items for one opportunity.
func (s *LineItemService) ListByOpportunity(ctx context.Context, opportunityID uint) ([]LineItem, error) {
	return s.repo.ListByOpportunity(ctx, opportunityID)
}
