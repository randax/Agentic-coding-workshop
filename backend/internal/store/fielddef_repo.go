package store

import (
	"context"

	"saltcrm/internal/studio"

	"gorm.io/gorm"
)

// FieldDefRepository is the GORM-backed implementation of studio.Repository.
type FieldDefRepository struct {
	db *gorm.DB
}

// NewFieldDefRepository wires a repository to a GORM database handle.
func NewFieldDefRepository(db *gorm.DB) *FieldDefRepository {
	return &FieldDefRepository{db: db}
}

// Create inserts a new custom-field definition.
func (r *FieldDefRepository) Create(ctx context.Context, d *studio.FieldDef) error {
	return r.db.WithContext(ctx).Create(d).Error
}

// ListByModule returns the custom-field definitions for a module, oldest first.
func (r *FieldDefRepository) ListByModule(ctx context.Context, module string) ([]studio.FieldDef, error) {
	var defs []studio.FieldDef
	if err := r.db.WithContext(ctx).Where("module = ?", module).Order("id").Find(&defs).Error; err != nil {
		return nil, err
	}
	return defs, nil
}
