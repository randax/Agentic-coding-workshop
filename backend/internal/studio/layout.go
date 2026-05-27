package studio

import (
	"context"
	"errors"
	"strings"
)

// ErrInvalidView is returned when a layout names a view other than the three
// generic views.
var ErrInvalidView = errors.New("invalid view")

// validViews are the views a layout can be saved for.
var validViews = map[string]bool{"list": true, "detail": true, "edit": true}

// LayoutDef is a saved view layout: the ordered, visible field names for one
// (module, view). It is unique on (module, view) so saving upserts.
type LayoutDef struct {
	ID     uint     `gorm:"primaryKey" json:"id"`
	Module string   `gorm:"uniqueIndex:idx_layout_module_view" json:"module"`
	View   string   `gorm:"uniqueIndex:idx_layout_module_view" json:"view"`
	Fields []string `gorm:"serializer:json" json:"fields"`
}

// LayoutRepository persists saved layouts. Upsert replaces the row for a
// (module, view); GetByModule returns every saved view for a module.
type LayoutRepository interface {
	GetByModule(ctx context.Context, module string) ([]LayoutDef, error)
	Upsert(ctx context.Context, d *LayoutDef) error
}

// SetLayout saves the ordered, visible field names for one view of a module,
// after validating the module and view. Saving replaces any existing layout for
// that (module, view).
func (s *Service) SetLayout(ctx context.Context, module, view string, fields []string) error {
	if strings.TrimSpace(module) == "" {
		return ErrModuleRequired
	}
	if !validViews[view] {
		return ErrInvalidView
	}
	return s.layouts.Upsert(ctx, &LayoutDef{Module: module, View: view, Fields: fields})
}

// GetLayouts returns a module's saved layouts keyed by view name. Views with no
// saved layout are absent.
func (s *Service) GetLayouts(ctx context.Context, module string) (map[string][]string, error) {
	defs, err := s.layouts.GetByModule(ctx, module)
	if err != nil {
		return nil, err
	}
	out := make(map[string][]string, len(defs))
	for _, d := range defs {
		out[d.View] = d.Fields
	}
	return out, nil
}
