# SaltCRM S12 — Studio: layout editing

- **Issue:** #43 (parent PRD #31). Blocked by #42 (custom fields), which is done.
- **User stories:** 49 (admin arranges a module's layout), 50 (changes persisted and applied for everyone).
- **Date:** 2026-05-27

## Goal

Let an admin choose **which fields/columns appear and in what order** for each
module's list, detail, and edit views. Selections persist to `layout_meta` and
the generic `/m/*` views render according to the saved layout for every user.

## Scope (decided)

- **Views covered:** list (columns), edit (form fields), and detail (fields).
- **Detail keeps code-defined panels.** The admin reorders/hides fields *within*
  the existing panel groupings; creating, renaming, or moving fields across
  panels is out of scope.
- **Modules:** all registered modules (layouts need no `custom_fields` column),
  so the Studio module picker lists every module in the metadata registry.
- **Editor UX:** each field row is a visibility checkbox plus Up/Down buttons.
  No drag-and-drop dependency; matches the existing plain-form style.

Out of scope: full panel-set editing (create/rename/reorder panels),
drag-and-drop, per-role or per-user layouts.

## Architecture

### Backend

#### 1. `metadata.Resolve` — the layout-resolution core (unit-tested per AC)

A pure function in the `metadata` package that replaces the metadata handler's
current `mergeCustomFields` helper:

```go
// Resolve merges runtime custom fields onto a module's code-defined metadata and
// applies saved view layouts. layouts maps a view name ("list"|"detail"|"edit")
// to its ordered, visible-only field names; a view absent from the map keeps its
// code-defined default. Resolve(base, custom, nil) reproduces the default
// (custom-merged) metadata, which is what the ?raw=1 endpoint serves.
func Resolve(base ModuleMeta, custom []Field, layouts map[string][]string) ModuleMeta
```

Behavior:

- **Custom merge (preserved):** append each custom field to `Fields` (flagged
  `Custom`); by default also append to list columns, edit fields, and a
  "Custom fields" detail panel — exactly today's behavior.
- **List:** if `layouts["list"]` is set, `ListView.Columns` becomes that list,
  filtered to field names that still exist. Otherwise the default columns.
- **Edit:** same against `EditView.Fields`.
- **Detail:** compute the *default merged panels* (code panels + the appended
  "Custom fields" panel). If `layouts["detail"]` is set, for each default panel
  in code order emit its fields filtered to the saved order, dropping panels that
  end up empty. Otherwise the default panels. Field→panel membership is fixed by
  code, so reordering is within-panel only.
- A field name in a saved layout that no longer exists is ignored (robust to
  fields removed after a layout was saved).

No import cycle: `Resolve` takes `custom []metadata.Field` (the handler converts
`studio.FieldDef` → `metadata.Field`, as it does today) and plain
`map[string][]string` layouts. `metadata` does not import `studio`.

#### 2. `studio` package gains layouts

The package already owns runtime configuration (custom field definitions); it
grows layout definitions alongside.

```go
// LayoutDef is a saved view layout: the ordered, visible field names for one
// (module, view). Unique on (module, view) so saving upserts.
type LayoutDef struct {
    ID     uint     `gorm:"primaryKey" json:"id"`
    Module string   `gorm:"uniqueIndex:idx_layout_module_view" json:"module"`
    View   string   `gorm:"uniqueIndex:idx_layout_module_view" json:"view"` // list|detail|edit
    Fields []string `gorm:"serializer:json" json:"fields"`
}
```

Service methods (validation lives here, not in the handler):

- `GetLayouts(ctx, module) (map[string][]string, error)` — view → ordered fields.
- `SetLayout(ctx, module, view string, fields []string) error` — validates
  `module` non-empty and `view ∈ {list,detail,edit}` (`ErrInvalidView`), then
  upserts the `(module, view)` row.

New `LayoutRepository` interface (mirrors `Repository` for field defs):

```go
type LayoutRepository interface {
    GetByModule(ctx context.Context, module string) ([]LayoutDef, error)
    Upsert(ctx context.Context, d *LayoutDef) error
}
```

`studio.NewService` grows a `layoutRepo LayoutRepository` parameter.

#### 3. Persistence

- `store/layoutdef_repo.go`: GORM-backed `LayoutDefRepository`
  (`GetByModule`, `Upsert` via `clause.OnConflict` on the unique index).
- `store.Migrate` adds `&studio.LayoutDef{}`.

#### 4. API

- `GET /metadata/{module}?raw=1` — base + custom fields, **layouts not applied**
  (the editor's design-time palette and defaults). Without `raw`, the handler
  fetches saved layouts and applies them.
- `GET /studio/layouts?module=X` (auth) → `{ "list": [...], "detail": [...],
  "edit": [...] }`; views with no saved layout are omitted.
- `PUT /studio/layouts` (admin only) → body `{ "module": "...", "views": {
  "list": [...], "detail": [...], "edit": [...] } }`; calls `SetLayout` for each
  provided view. Returns 204.

Wiring: the metadata handler reaches layouts through the existing `studioSvc`
reference (no new constructor arg). New routes register on the existing
`studioGroup` (auth), with the PUT gated by `requireRole(agent.RoleAdmin)`,
mirroring `POST /studio/fields`.

### Frontend

- `lib/api.ts`:
  - `getModuleMeta(module, cookie?, opts?: { raw?: boolean })` — appends `?raw=1`.
  - `getLayouts(module, cookie?) → { list?: string[]; detail?: string[]; edit?: string[] }`.
  - `saveLayouts(module, views) → void` (PUT, `credentials:"include"`).
- `app/studio/LayoutEditor.tsx` — client component, same pattern as `StudioForm`:
  - Props: `module: string` (re-fetches when it changes).
  - Fetches raw `ModuleMeta` (palette + per-view defaults + panel grouping) and
    saved layouts; initial selection per view = `saved[view] ?? rawDefault[view]`.
  - Renders three sections — **List columns** (flat), **Detail fields** (grouped
    under read-only code panel headers; Up/Down reorders within a panel), **Edit
    fields** (flat) — each row a checkbox + Up/Down.
  - One "Save layout" button → `saveLayouts(module, { list, detail, edit })`,
    sending each view's visible names in order. Confirms on success, shows an
    error on failure (admin required), and `router.refresh()`.
- `app/studio/page.tsx` — keep the custom-fields `StudioForm` (accounts); add a
  module `<select>` (all registered modules) bound to a `LayoutEditor`.

## Testing

- **`metadata` (unit, AC):** list-column layout applied; edit fields hidden +
  reordered; detail reordered within panels (and emptied panel dropped);
  no-layout fallback to defaults; custom-field merge preserved
  (`Resolve(base, custom, nil)`).
- **`studio` (unit, fake repo):** `SetLayout` rejects an invalid view
  (`ErrInvalidView`) and a missing module; `GetLayouts` returns saved views.
- **API (integration, throwaway SQLite):** `PUT /studio/layouts` is 403 for a
  manager and 204 for an admin; after an admin saves a list layout,
  `GET /metadata/accounts` reflects the new column order; `?raw=1` ignores the
  saved layout.
- **Frontend (component, mocked api — AC):** `LayoutEditor` renders fields with
  checkboxes from a mocked raw meta + layouts; toggling a checkbox hides a field;
  Up/Down reorders; Save calls `saveLayouts` with the expected per-view payload.

## Delivery

One vertical slice / PR (S12 is not flagged HITL). May be split into S12a
(backend resolution + API) and S12b (frontend editor) if smaller PRs are
preferred. Built test-first, one behavior at a time, shipping green.
