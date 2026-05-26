package store

import (
	"context"
	"errors"

	"saltcrm/internal/customer"

	"gorm.io/gorm"
)

// CustomerRepository is the GORM-backed implementation of customer.Repository.
type CustomerRepository struct {
	db *gorm.DB
}

// NewCustomerRepository wires a repository to a GORM database handle.
func NewCustomerRepository(db *gorm.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

// List returns all customers ordered by name.
func (r *CustomerRepository) List(ctx context.Context) ([]customer.Customer, error) {
	var customers []customer.Customer
	if err := r.db.WithContext(ctx).Order("name").Find(&customers).Error; err != nil {
		return nil, err
	}
	return customers, nil
}

// Get returns a single customer by ID, translating GORM's not-found error into
// the domain-level customer.ErrNotFound.
func (r *CustomerRepository) Get(ctx context.Context, id uint) (customer.Customer, error) {
	var c customer.Customer
	if err := r.db.WithContext(ctx).First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customer.Customer{}, customer.ErrNotFound
		}
		return customer.Customer{}, err
	}
	return c, nil
}

// Create inserts a new customer.
func (r *CustomerRepository) Create(ctx context.Context, c *customer.Customer) error {
	return r.db.WithContext(ctx).Create(c).Error
}

// Update persists all fields of an existing customer.
func (r *CustomerRepository) Update(ctx context.Context, c *customer.Customer) error {
	return r.db.WithContext(ctx).Save(c).Error
}
