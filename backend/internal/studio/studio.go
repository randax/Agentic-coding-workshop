// Package studio owns runtime custom-field definitions. Admins add custom
// fields to a module; the definitions live in fields_meta and are merged into a
// module's metadata so the generic views render them, while the values live in
// each record's custom_fields JSON. The service depends on the Repository
// interface, not a database, so it is unit-testable in isolation.
package studio

import (
	"context"
	"errors"
	"strings"
)

// Validation errors.
var (
	ErrModuleRequired = errors.New("module is required")
	ErrNameRequired   = errors.New("field name is required")
	ErrLabelRequired  = errors.New("field label is required")
	ErrInvalidType    = errors.New("invalid field type")
)

// validTypes are the field types Studio can create (a subset of the metadata
// field types, kept as strings to avoid importing the metadata package).
var validTypes = map[string]bool{
	"string": true, "enum": true, "currency": true, "bool": true, "date": true,
}

// FieldDef is a runtime-defined custom field on a module.
type FieldDef struct {
	ID      uint     `gorm:"primaryKey" json:"id"`
	Module  string   `gorm:"index" json:"module"`
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Label   string   `json:"label"`
	Options []string `gorm:"serializer:json" json:"options,omitempty"`
}

// Repository is the persistence seam the service depends on.
type Repository interface {
	Create(ctx context.Context, d *FieldDef) error
	ListByModule(ctx context.Context, module string) ([]FieldDef, error)
}

// Service owns custom-field-definition logic.
type Service struct {
	repo Repository
}

// NewService wires a Service to its repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// AddField defines a new custom field on a module, after validation.
func (s *Service) AddField(ctx context.Context, d FieldDef) (FieldDef, error) {
	if strings.TrimSpace(d.Module) == "" {
		return FieldDef{}, ErrModuleRequired
	}
	if strings.TrimSpace(d.Name) == "" {
		return FieldDef{}, ErrNameRequired
	}
	if strings.TrimSpace(d.Label) == "" {
		return FieldDef{}, ErrLabelRequired
	}
	if !validTypes[d.Type] {
		return FieldDef{}, ErrInvalidType
	}
	d.ID = 0
	if err := s.repo.Create(ctx, &d); err != nil {
		return FieldDef{}, err
	}
	return d, nil
}

// ListByModule returns the custom field definitions for a module.
func (s *Service) ListByModule(ctx context.Context, module string) ([]FieldDef, error) {
	return s.repo.ListByModule(ctx, module)
}
