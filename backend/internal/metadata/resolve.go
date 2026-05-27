package metadata

import "sort"

// CustomFieldsPanel is the detail-view panel label under which runtime custom
// fields appear by default.
const CustomFieldsPanel = "Custom fields"

// Resolve produces the final ModuleMeta served to the generic views: it merges
// runtime custom fields onto a module's code-defined metadata and applies any
// saved view layouts.
//
// custom holds the module's custom fields (already converted from Studio
// definitions). layouts maps a view name ("list"|"detail"|"edit") to its
// ordered, visible-only field names; a view absent from the map keeps its
// code-defined default. Field names in a layout that no longer exist are
// ignored. Resolve(base, custom, nil) reproduces the default custom-merged
// metadata — what GET /metadata/{module}?raw=1 serves.
//
// Resolve never mutates base (the registry shares it across requests): every
// slice in the result is freshly allocated.
func Resolve(base ModuleMeta, custom []Field, layouts map[string][]string) ModuleMeta {
	out := base // copies scalar fields and passes Subpanels through

	// Merged field set: code fields then custom fields.
	out.Fields = append(append(make([]Field, 0, len(base.Fields)+len(custom)), base.Fields...), custom...)

	customNames := make([]string, len(custom))
	for i, f := range custom {
		customNames[i] = f.Name
	}
	exists := make(map[string]bool, len(out.Fields))
	for _, f := range out.Fields {
		exists[f.Name] = true
	}

	// List & edit default to the code layout plus custom fields appended; a saved
	// layout overrides the order and visibility.
	defaultList := append(append([]string{}, base.ListView.Columns...), customNames...)
	out.ListView = ListView{Columns: applyLayout(layouts["list"], hasLayout(layouts, "list"), defaultList, exists)}

	defaultEdit := append(append([]string{}, base.EditView.Fields...), customNames...)
	out.EditView = EditView{Fields: applyLayout(layouts["edit"], hasLayout(layouts, "edit"), defaultEdit, exists)}

	out.DetailView = resolveDetail(base.DetailView, customNames, layouts, exists)

	return out
}

func hasLayout(layouts map[string][]string, view string) bool {
	_, ok := layouts[view]
	return ok
}

// applyLayout returns the saved order (filtered to existing fields) when a
// layout is present, otherwise a fresh copy of the default.
func applyLayout(saved []string, present bool, def []string, exists map[string]bool) []string {
	if !present {
		return append([]string{}, def...)
	}
	out := make([]string, 0, len(saved))
	for _, name := range saved {
		if exists[name] {
			out = append(out, name)
		}
	}
	return out
}

// resolveDetail keeps code-defined panel membership: each field stays in its
// panel; a saved "detail" layout only reorders fields within their panel and
// hides the rest, dropping panels left empty.
func resolveDetail(base DetailView, customNames []string, layouts map[string][]string, exists map[string]bool) DetailView {
	// Default panels = code panels + a Custom fields panel.
	panels := append([]Panel{}, base.Panels...)
	if len(customNames) > 0 {
		panels = append(panels, Panel{Label: CustomFieldsPanel, Fields: customNames})
	}

	saved, present := layouts["detail"]
	if !present {
		out := make([]Panel, len(panels))
		for i, p := range panels {
			out[i] = Panel{Label: p.Label, Fields: append([]string{}, p.Fields...)}
		}
		return DetailView{Panels: out}
	}

	order := make(map[string]int, len(saved))
	for i, name := range saved {
		if exists[name] {
			order[name] = i
		}
	}
	var out []Panel
	for _, p := range panels {
		var fields []string
		for _, name := range p.Fields {
			if _, shown := order[name]; shown {
				fields = append(fields, name)
			}
		}
		sort.SliceStable(fields, func(a, b int) bool { return order[fields[a]] < order[fields[b]] })
		if len(fields) > 0 {
			out = append(out, Panel{Label: p.Label, Fields: fields})
		}
	}
	return DetailView{Panels: out}
}
