package store

import (
	"context"

	"saltcrm/internal/studio"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// LayoutDefRepository is the GORM-backed implementation of studio.LayoutRepository.
type LayoutDefRepository struct {
	db *gorm.DB
}

// NewLayoutDefRepository wires a repository to a GORM database handle.
func NewLayoutDefRepository(db *gorm.DB) *LayoutDefRepository {
	return &LayoutDefRepository{db: db}
}

// Upsert saves a layout, replacing the existing row for the same (module, view).
func (r *LayoutDefRepository) Upsert(ctx context.Context, d *studio.LayoutDef) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "module"}, {Name: "view"}},
		DoUpdates: clause.AssignmentColumns([]string{"fields"}),
	}).Create(d).Error
}

// GetByModule returns the saved layouts for a module, oldest first.
func (r *LayoutDefRepository) GetByModule(ctx context.Context, module string) ([]studio.LayoutDef, error) {
	var defs []studio.LayoutDef
	if err := r.db.WithContext(ctx).Where("module = ?", module).Order("id").Find(&defs).Error; err != nil {
		return nil, err
	}
	return defs, nil
}
