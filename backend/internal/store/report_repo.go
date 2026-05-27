package store

import (
	"context"
	"errors"

	"saltcrm/internal/report"

	"gorm.io/gorm"
)

// ReportRepository is the GORM-backed implementation of report.Repository.
type ReportRepository struct {
	db *gorm.DB
}

// NewReportRepository wires a repository to a GORM database handle.
func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

// Create inserts a new saved report.
func (r *ReportRepository) Create(ctx context.Context, s *report.Saved) error {
	return r.db.WithContext(ctx).Create(s).Error
}

// List returns all saved reports, oldest first.
func (r *ReportRepository) List(ctx context.Context) ([]report.Saved, error) {
	var reports []report.Saved
	if err := r.db.WithContext(ctx).Order("id").Find(&reports).Error; err != nil {
		return nil, err
	}
	return reports, nil
}

// Get returns a saved report by ID, translating GORM's not-found error into the
// domain-level report.ErrNotFound.
func (r *ReportRepository) Get(ctx context.Context, id uint) (report.Saved, error) {
	var s report.Saved
	if err := r.db.WithContext(ctx).First(&s, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return report.Saved{}, report.ErrNotFound
		}
		return report.Saved{}, err
	}
	return s, nil
}
