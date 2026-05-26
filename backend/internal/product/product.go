// Package product holds the product-catalog domain model and the service that
// owns catalog business logic (creating, editing, listing, and retiring the
// offerings the ISP sells). The service depends on the Repository interface,
// not on any database, so it can be unit-tested in isolation.
package product

import (
	"context"
	"errors"
)

// Category is the kind of product the ISP sells.
type Category string

const (
	CategoryFiber  Category = "fiber"
	CategoryRouter Category = "router"
	CategoryTV     Category = "tv"
)

// Valid reports whether c is a known product category.
func (c Category) Valid() bool {
	switch c {
	case CategoryFiber, CategoryRouter, CategoryTV:
		return true
	default:
		return false
	}
}

// ErrNotFound is returned when a requested product does not exist.
var ErrNotFound = errors.New("product not found")

// ErrInvalidCategory is returned when a product has an unknown category.
var ErrInvalidCategory = errors.New("invalid product category")

// ErrInvalidPrice is returned when a product's monthly price is negative.
var ErrInvalidPrice = errors.New("product price must not be negative")

// Product is a catalog offering. Common fields apply to every product; the
// per-category attributes are nullable and only set for their category
// (SpeedMbps for fiber, RouterModel for routers, TvPackageTier for TV).
type Product struct {
	ID           uint     `gorm:"primaryKey" json:"id"`
	Name         string   `json:"name"`
	Category     Category `json:"category"`
	MonthlyPrice float64  `json:"monthlyPrice"`
	Available    bool     `json:"available"`

	SpeedMbps     *int    `json:"speedMbps,omitempty"`
	RouterModel   *string `json:"routerModel,omitempty"`
	TvPackageTier *string `json:"tvPackageTier,omitempty"`
}

// Repository is the persistence seam the catalog service depends on.
type Repository interface {
	List(ctx context.Context) ([]Product, error)
	Get(ctx context.Context, id uint) (Product, error)
	Create(ctx context.Context, p *Product) error
	Update(ctx context.Context, p *Product) error
}

// Service owns catalog business logic.
type Service struct {
	repo Repository
}

// NewService wires a Service to its repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// List returns all products in the catalog.
func (s *Service) List(ctx context.Context) ([]Product, error) {
	return s.repo.List(ctx)
}

// Get returns a single product by ID, or ErrNotFound.
func (s *Service) Get(ctx context.Context, id uint) (Product, error) {
	return s.repo.Get(ctx, id)
}

// Update edits an existing product, applying the same validation as Create.
func (s *Service) Update(ctx context.Context, p Product) (Product, error) {
	if !p.Category.Valid() {
		return Product{}, ErrInvalidCategory
	}
	if p.MonthlyPrice < 0 {
		return Product{}, ErrInvalidPrice
	}
	if err := s.repo.Update(ctx, &p); err != nil {
		return Product{}, err
	}
	return p, nil
}

// Retire marks a product unavailable so it can no longer be subscribed to.
func (s *Service) Retire(ctx context.Context, id uint) error {
	p, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	p.Available = false
	return s.repo.Update(ctx, &p)
}

// Create adds a new product to the catalog. New products are available.
func (s *Service) Create(ctx context.Context, p Product) (Product, error) {
	if !p.Category.Valid() {
		return Product{}, ErrInvalidCategory
	}
	if p.MonthlyPrice < 0 {
		return Product{}, ErrInvalidPrice
	}
	p.Available = true
	if err := s.repo.Create(ctx, &p); err != nil {
		return Product{}, err
	}
	return p, nil
}
