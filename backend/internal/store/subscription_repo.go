package store

import (
	"context"

	"ispcrm/internal/subscription"

	"gorm.io/gorm"
)

// SubscriptionRepository is the GORM-backed implementation of
// subscription.Repository.
type SubscriptionRepository struct {
	db *gorm.DB
}

// NewSubscriptionRepository wires a repository to a GORM database handle.
func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// ListByCustomer returns a customer's subscriptions, with each subscription's
// catalog product preloaded for display.
func (r *SubscriptionRepository) ListByCustomer(ctx context.Context, customerID uint) ([]subscription.Subscription, error) {
	var subs []subscription.Subscription
	if err := r.db.WithContext(ctx).
		Preload("Product").
		Where("customer_id = ?", customerID).
		Order("start_date desc").
		Find(&subs).Error; err != nil {
		return nil, err
	}
	return subs, nil
}
