package store

import (
	"context"
	"errors"

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

// Get returns a single subscription by ID with its catalog product preloaded,
// translating GORM's not-found error into the domain-level ErrNotFound.
func (r *SubscriptionRepository) Get(ctx context.Context, id uint) (subscription.Subscription, error) {
	var sub subscription.Subscription
	if err := r.db.WithContext(ctx).Preload("Product").First(&sub, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return subscription.Subscription{}, subscription.ErrNotFound
		}
		return subscription.Subscription{}, err
	}
	return sub, nil
}

// Create inserts a new subscription.
func (r *SubscriptionRepository) Create(ctx context.Context, s *subscription.Subscription) error {
	return r.db.WithContext(ctx).Create(s).Error
}

// Update persists all fields of an existing subscription. It omits the Product
// association so saving a subscription never writes back to the catalog (the
// snapshot price lives on the subscription row itself).
func (r *SubscriptionRepository) Update(ctx context.Context, s *subscription.Subscription) error {
	return r.db.WithContext(ctx).Omit("Product").Save(s).Error
}
