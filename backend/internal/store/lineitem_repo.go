package store

import (
	"context"

	"saltcrm/internal/opportunity"

	"gorm.io/gorm"
)

// LineItemRepository is the GORM-backed implementation of
// opportunity.LineItemRepository.
type LineItemRepository struct {
	db *gorm.DB
}

// NewLineItemRepository wires a repository to a GORM database handle.
func NewLineItemRepository(db *gorm.DB) *LineItemRepository {
	return &LineItemRepository{db: db}
}

// Create inserts a new line item.
func (r *LineItemRepository) Create(ctx context.Context, li *opportunity.LineItem) error {
	return r.db.WithContext(ctx).Create(li).Error
}

// ListByOpportunity returns the line items for one opportunity, oldest first.
func (r *LineItemRepository) ListByOpportunity(ctx context.Context, opportunityID uint) ([]opportunity.LineItem, error) {
	var items []opportunity.LineItem
	if err := r.db.WithContext(ctx).Where("opportunity_id = ?", opportunityID).Order("id").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
