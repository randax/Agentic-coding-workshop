package store

import (
	"context"
	"errors"

	"saltcrm/internal/contact"

	"gorm.io/gorm"
)

// ContactRepository is the GORM-backed implementation of contact.Repository.
type ContactRepository struct {
	db *gorm.DB
}

// NewContactRepository wires a repository to a GORM database handle.
func NewContactRepository(db *gorm.DB) *ContactRepository {
	return &ContactRepository{db: db}
}

// List returns all contacts ordered by name.
func (r *ContactRepository) List(ctx context.Context) ([]contact.Contact, error) {
	var contacts []contact.Contact
	if err := r.db.WithContext(ctx).Order("name").Find(&contacts).Error; err != nil {
		return nil, err
	}
	return contacts, nil
}

// ListByAccount returns the contacts for one account, ordered by name.
func (r *ContactRepository) ListByAccount(ctx context.Context, accountID uint) ([]contact.Contact, error) {
	var contacts []contact.Contact
	if err := r.db.WithContext(ctx).Where("account_id = ?", accountID).Order("name").Find(&contacts).Error; err != nil {
		return nil, err
	}
	return contacts, nil
}

// Get returns a single contact by ID, translating GORM's not-found error.
func (r *ContactRepository) Get(ctx context.Context, id uint) (contact.Contact, error) {
	var c contact.Contact
	if err := r.db.WithContext(ctx).First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return contact.Contact{}, contact.ErrNotFound
		}
		return contact.Contact{}, err
	}
	return c, nil
}

// Create inserts a new contact.
func (r *ContactRepository) Create(ctx context.Context, c *contact.Contact) error {
	return r.db.WithContext(ctx).Create(c).Error
}

// Update persists all fields of an existing contact.
func (r *ContactRepository) Update(ctx context.Context, c *contact.Contact) error {
	return r.db.WithContext(ctx).Save(c).Error
}
