package store

import (
	"context"
	"errors"

	"saltcrm/internal/opportunity"

	"gorm.io/gorm"
)

// OpportunityRepository is the GORM-backed implementation of opportunity.Repository.
type OpportunityRepository struct {
	db *gorm.DB
}

// NewOpportunityRepository wires a repository to a GORM database handle.
func NewOpportunityRepository(db *gorm.DB) *OpportunityRepository {
	return &OpportunityRepository{db: db}
}

// List returns all opportunities ordered by expected close date.
func (r *OpportunityRepository) List(ctx context.Context) ([]opportunity.Opportunity, error) {
	var opps []opportunity.Opportunity
	if err := r.db.WithContext(ctx).Order("expected_close_date").Find(&opps).Error; err != nil {
		return nil, err
	}
	return opps, nil
}

// Get returns a single opportunity by ID, translating GORM's not-found error.
func (r *OpportunityRepository) Get(ctx context.Context, id uint) (opportunity.Opportunity, error) {
	var o opportunity.Opportunity
	if err := r.db.WithContext(ctx).First(&o, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return opportunity.Opportunity{}, opportunity.ErrNotFound
		}
		return opportunity.Opportunity{}, err
	}
	return o, nil
}

// Create inserts a new opportunity.
func (r *OpportunityRepository) Create(ctx context.Context, o *opportunity.Opportunity) error {
	return r.db.WithContext(ctx).Create(o).Error
}

// Update persists all fields of an existing opportunity.
func (r *OpportunityRepository) Update(ctx context.Context, o *opportunity.Opportunity) error {
	return r.db.WithContext(ctx).Save(o).Error
}
