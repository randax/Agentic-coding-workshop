package store

import (
	"context"
	"errors"

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

// Create inserts a new case.
func (r *CaseRepository) Create(ctx context.Context, c *supportcase.Case) error {
	return r.db.WithContext(ctx).Create(c).Error
}

// CreateComment appends a comment to a case's timeline.
func (r *CaseRepository) CreateComment(ctx context.Context, cm *supportcase.CaseComment) error {
	return r.db.WithContext(ctx).Create(cm).Error
}

// Get returns a single case with its assigned agent and its comment timeline
// (oldest first, each comment's author preloaded), translating GORM's not-found
// error into the domain-level supportcase.ErrNotFound.
func (r *CaseRepository) Get(ctx context.Context, id uint) (supportcase.Case, error) {
	var c supportcase.Case
	if err := r.db.WithContext(ctx).
		Preload("AssignedAgent").
		Preload("Comments", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at asc")
		}).
		Preload("Comments.AuthorAgent").
		First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return supportcase.Case{}, supportcase.ErrNotFound
		}
		return supportcase.Case{}, err
	}
	return c, nil
}
