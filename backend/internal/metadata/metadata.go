// Package metadata describes SaltCRM modules as data: the fields a module has,
// how they're typed and labelled, and how its list/detail/edit views are laid
// out. The generic frontend renders any module from this metadata, so adding a
// field or module needs no new screen code. This package holds the types and a
// registry; module definitions are registered by the wiring layer.
package metadata

import (
	"errors"
	"fmt"
)

// FieldType classifies how a field is rendered and formatted by the generic UI.
type FieldType string

const (
	FieldString   FieldType = "string"
	FieldEnum     FieldType = "enum"
	FieldCurrency FieldType = "currency"
	FieldBool     FieldType = "bool"
	FieldDate     FieldType = "date"
)

// Field describes a single field on a module.
type Field struct {
	Name    string    `json:"name"`
	Type    FieldType `json:"type"`
	Label   string    `json:"label"`
	Options []string  `json:"options,omitempty"` // for FieldEnum
	Custom  bool      `json:"custom,omitempty"`  // runtime-defined (Studio) field
}

// ListView is the column layout of a module's list view.
type ListView struct {
	Columns []string `json:"columns"`
}

// DetailView is the panel layout of a module's record view (stubbed for now).
type DetailView struct {
	Panels []Panel `json:"panels"`
}

// Panel groups fields on a detail view.
type Panel struct {
	Label  string   `json:"label"`
	Fields []string `json:"fields"`
}

// EditView is the field layout of a module's edit form (stubbed for now).
type EditView struct {
	Fields []string `json:"fields"`
}

// Action is a button on a record view that POSTs to an endpoint (with "{id}"
// replaced by the record's id) — e.g. converting a lead. Generic so any module
// can declare record-level actions without bespoke UI.
type Action struct {
	Label  string `json:"label"`
	Method string `json:"method"`
	Path   string `json:"path"`
}

// Subpanel describes a related-records panel on a record view. It is
// self-contained: Path is the records endpoint (with "{id}" replaced by the
// parent record's id) and Columns are the fields to show, so a subpanel renders
// without depending on another registered module.
type Subpanel struct {
	Label   string  `json:"label"`
	Path    string  `json:"path"`
	Columns []Field `json:"columns"`
}

// ModuleMeta is the complete metadata for one module — the payload served at
// GET /metadata/{module} and consumed by the generic views.
type ModuleMeta struct {
	Module        string     `json:"module"`
	Label         string     `json:"label"`
	LabelSingular string     `json:"labelSingular"`
	Fields        []Field    `json:"fields"`
	ListView      ListView   `json:"listView"`
	DetailView    DetailView `json:"detailView"`
	EditView      EditView   `json:"editView"`
	Subpanels     []Subpanel `json:"subpanels"`
	Actions       []Action   `json:"actions,omitempty"`
}

// ErrUnknownModule is returned by Registry.Get for a module that was never registered.
var ErrUnknownModule = errors.New("unknown module")

// Registry holds the metadata for every registered module.
type Registry struct {
	modules map[string]ModuleMeta
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{modules: map[string]ModuleMeta{}}
}

// Register adds (or replaces) a module's metadata.
func (r *Registry) Register(m ModuleMeta) {
	r.modules[m.Module] = m
}

// Get returns a module's metadata, or ErrUnknownModule.
func (r *Registry) Get(module string) (ModuleMeta, error) {
	m, ok := r.modules[module]
	if !ok {
		return ModuleMeta{}, fmt.Errorf("%q: %w", module, ErrUnknownModule)
	}
	return m, nil
}
