package store

import (
	"context"

	"ispcrm/internal/customer"

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
