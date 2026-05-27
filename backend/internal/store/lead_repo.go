package store

import (
	"context"
	"errors"

	"saltcrm/internal/lead"

	"gorm.io/gorm"
)

// LeadRepository is the GORM-backed implementation of lead.Repository.
type LeadRepository struct {
	db *gorm.DB
}

// NewLeadRepository wires a repository to a GORM database handle.
func NewLeadRepository(db *gorm.DB) *LeadRepository {
	return &LeadRepository{db: db}
}

// List returns all leads ordered by name.
func (r *LeadRepository) List(ctx context.Context) ([]lead.Lead, error) {
	var leads []lead.Lead
	if err := r.db.WithContext(ctx).Order("name").Find(&leads).Error; err != nil {
		return nil, err
	}
	return leads, nil
}

// Get returns a single lead by ID, translating GORM's not-found error.
func (r *LeadRepository) Get(ctx context.Context, id uint) (lead.Lead, error) {
	var l lead.Lead
	if err := r.db.WithContext(ctx).First(&l, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return lead.Lead{}, lead.ErrNotFound
		}
		return lead.Lead{}, err
	}
	return l, nil
}

// Create inserts a new lead.
func (r *LeadRepository) Create(ctx context.Context, l *lead.Lead) error {
	return r.db.WithContext(ctx).Create(l).Error
}

// Update persists all fields of an existing lead.
func (r *LeadRepository) Update(ctx context.Context, l *lead.Lead) error {
	return r.db.WithContext(ctx).Save(l).Error
}
