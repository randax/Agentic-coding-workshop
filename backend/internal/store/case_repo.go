package store

import (
	"context"

	"ispcrm/internal/supportcase"

	"gorm.io/gorm"
)

// CaseRepository is the GORM-backed implementation of supportcase.Repository.
type CaseRepository struct {
	db *gorm.DB
}

// NewCaseRepository wires a repository to a GORM database handle.
func NewCaseRepository(db *gorm.DB) *CaseRepository {
	return &CaseRepository{db: db}
}

// ListByCustomer returns a customer's cases, newest first, with the assigned
// agent preloaded for display.
func (r *CaseRepository) ListByCustomer(ctx context.Context, customerID uint) ([]supportcase.Case, error) {
	var cases []supportcase.Case
	if err := r.db.WithContext(ctx).
		Preload("AssignedAgent").
		Where("customer_id = ?", customerID).
		Order("created_at desc").
		Find(&cases).Error; err != nil {
		return nil, err
	}
	return cases, nil
}
