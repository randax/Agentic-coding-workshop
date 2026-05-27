package store

import (
	"context"
	"errors"

	"saltcrm/internal/activity"

	"gorm.io/gorm"
)

// ActivityRepository is the GORM-backed implementation of activity.Repository.
type ActivityRepository struct {
	db *gorm.DB
}

// NewActivityRepository wires a repository to a GORM database handle.
func NewActivityRepository(db *gorm.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

// List returns all activities, most recent first.
func (r *ActivityRepository) List(ctx context.Context) ([]activity.Activity, error) {
	var items []activity.Activity
	if err := r.db.WithContext(ctx).Order("occurred_at desc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// ListForParent returns the activities logged against one record, most recent first.
func (r *ActivityRepository) ListForParent(ctx context.Context, parentType string, parentID uint) ([]activity.Activity, error) {
	var items []activity.Activity
	if err := r.db.WithContext(ctx).
		Where("parent_type = ? AND parent_id = ?", parentType, parentID).
		Order("occurred_at desc").
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// Get returns a single activity by ID, translating GORM's not-found error.
func (r *ActivityRepository) Get(ctx context.Context, id uint) (activity.Activity, error) {
	var a activity.Activity
	if err := r.db.WithContext(ctx).First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return activity.Activity{}, activity.ErrNotFound
		}
		return activity.Activity{}, err
	}
	return a, nil
}

// Create inserts a new activity.
func (r *ActivityRepository) Create(ctx context.Context, a *activity.Activity) error {
	return r.db.WithContext(ctx).Create(a).Error
}

// Update persists all fields of an existing activity.
func (r *ActivityRepository) Update(ctx context.Context, a *activity.Activity) error {
	return r.db.WithContext(ctx).Save(a).Error
}
