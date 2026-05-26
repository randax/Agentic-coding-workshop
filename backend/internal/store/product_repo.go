package store

import (
	"context"
	"errors"

	"ispcrm/internal/product"

	"gorm.io/gorm"
)

// ProductRepository is the GORM-backed implementation of product.Repository.
type ProductRepository struct {
	db *gorm.DB
}

// NewProductRepository wires a repository to a GORM database handle.
func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// List returns all products ordered by category then name.
func (r *ProductRepository) List(ctx context.Context) ([]product.Product, error) {
	var products []product.Product
	if err := r.db.WithContext(ctx).Order("category, name").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// Get returns a single product by ID, translating GORM's not-found error into
// the domain-level product.ErrNotFound.
func (r *ProductRepository) Get(ctx context.Context, id uint) (product.Product, error) {
	var p product.Product
	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return product.Product{}, product.ErrNotFound
		}
		return product.Product{}, err
	}
	return p, nil
}

// Create inserts a new product.
func (r *ProductRepository) Create(ctx context.Context, p *product.Product) error {
	return r.db.WithContext(ctx).Create(p).Error
}

// Update persists all fields of an existing product.
func (r *ProductRepository) Update(ctx context.Context, p *product.Product) error {
	return r.db.WithContext(ctx).Save(p).Error
}
