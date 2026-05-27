package metadata

import (
	"reflect"
	"testing"
)

// resolveBase is a small two-panel module used across the Resolve tests.
func resolveBase() ModuleMeta {
	return ModuleMeta{
		Module: "accounts", Label: "Accounts", LabelSingular: "Account",
		Fields: []Field{
			{Name: "name", Type: FieldString, Label: "Name"},
			{Name: "email", Type: FieldString, Label: "Email"},
			{Name: "phone", Type: FieldString, Label: "Phone"},
			{Name: "status", Type: FieldEnum, Label: "Status", Options: []string{"active", "suspended"}},
		},
		ListView: ListView{Columns: []string{"name", "email", "status"}},
		DetailView: DetailView{Panels: []Panel{
			{Label: "Profile", Fields: []string{"name", "email", "phone"}},
			{Label: "Account", Fields: []string{"status"}},
		}},
		EditView: EditView{Fields: []string{"name", "email", "phone", "status"}},
	}
}

func resolveCustom() []Field {
	return []Field{{Name: "churnRisk", Type: FieldEnum, Label: "Churn risk", Options: []string{"low", "high"}, Custom: true}}
}

func fieldNames(fs []Field) []string {
	out := make([]string, len(fs))
	for i, f := range fs {
		out[i] = f.Name
	}
	return out
}

func TestResolveMergesCustomFieldsByDefault(t *testing.T) {
	m := Resolve(resolveBase(), resolveCustom(), nil)

	// Custom field is appended to the field set, flagged custom.
	var churn *Field
	for i := range m.Fields {
		if m.Fields[i].Name == "churnRisk" {
			churn = &m.Fields[i]
		}
	}
	if churn == nil || !churn.Custom || churn.Type != FieldEnum {
		t.Fatalf("churnRisk not merged as a custom enum field: %+v", m.Fields)
	}
	// It appears in list columns and edit fields by default.
	if got := m.ListView.Columns; !reflect.DeepEqual(got, []string{"name", "email", "status", "churnRisk"}) {
		t.Errorf("list columns = %v, want default + churnRisk appended", got)
	}
	if got := m.EditView.Fields; got[len(got)-1] != "churnRisk" {
		t.Errorf("edit fields = %v, want churnRisk appended", got)
	}
	// And gets its own detail panel.
	last := m.DetailView.Panels[len(m.DetailView.Panels)-1]
	if last.Label != "Custom fields" || !reflect.DeepEqual(last.Fields, []string{"churnRisk"}) {
		t.Errorf("last detail panel = %+v, want Custom fields/[churnRisk]", last)
	}
}

func TestResolveAppliesListColumnLayout(t *testing.T) {
	layouts := map[string][]string{"list": {"status", "name"}}

	m := Resolve(resolveBase(), resolveCustom(), layouts)

	if got := m.ListView.Columns; !reflect.DeepEqual(got, []string{"status", "name"}) {
		t.Errorf("list columns = %v, want [status name] (reordered, others hidden)", got)
	}
}

func TestResolveHidesAndReordersEditFields(t *testing.T) {
	layouts := map[string][]string{"edit": {"name", "churnRisk"}}

	m := Resolve(resolveBase(), resolveCustom(), layouts)

	if got := m.EditView.Fields; !reflect.DeepEqual(got, []string{"name", "churnRisk"}) {
		t.Errorf("edit fields = %v, want [name churnRisk]", got)
	}
}

func TestResolveReordersDetailWithinPanels(t *testing.T) {
	// Saved order lists email before name; phone and status hidden. churnRisk shown.
	layouts := map[string][]string{"detail": {"email", "name", "churnRisk"}}

	m := Resolve(resolveBase(), resolveCustom(), layouts)

	// Account panel (only status) is emptied and dropped. Profile reorders to
	// [email, name] (within-panel only — churnRisk stays in Custom fields).
	want := []Panel{
		{Label: "Profile", Fields: []string{"email", "name"}},
		{Label: "Custom fields", Fields: []string{"churnRisk"}},
	}
	if !reflect.DeepEqual(m.DetailView.Panels, want) {
		t.Errorf("detail panels = %+v, want %+v", m.DetailView.Panels, want)
	}
}

func TestResolveIgnoresUnknownLayoutFields(t *testing.T) {
	layouts := map[string][]string{"list": {"name", "ghost", "status"}}

	m := Resolve(resolveBase(), resolveCustom(), layouts)

	if got := m.ListView.Columns; !reflect.DeepEqual(got, []string{"name", "status"}) {
		t.Errorf("list columns = %v, want [name status] (ghost dropped)", got)
	}
}

func TestResolveDoesNotMutateBase(t *testing.T) {
	base := resolveBase()
	Resolve(base, resolveCustom(), map[string][]string{"list": {"status"}, "detail": {"name"}})

	if got := base.ListView.Columns; !reflect.DeepEqual(got, []string{"name", "email", "status"}) {
		t.Errorf("base list columns mutated: %v", got)
	}
	if got := fieldNames(base.Fields); !reflect.DeepEqual(got, []string{"name", "email", "phone", "status"}) {
		t.Errorf("base fields mutated: %v", got)
	}
	if got := base.DetailView.Panels[0].Fields; !reflect.DeepEqual(got, []string{"name", "email", "phone"}) {
		t.Errorf("base detail panel mutated: %v", got)
	}
}
