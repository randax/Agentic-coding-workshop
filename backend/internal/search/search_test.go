package search

import "testing"

// findGroup returns the group for a module, or nil if absent.
func findGroup(groups []Group, m Module) *Group {
	for i := range groups {
		if groups[i].Module == m {
			return &groups[i]
		}
	}
	return nil
}

func TestSearchRanksTitleMatchAboveFieldMatch(t *testing.T) {
	candidates := []Candidate{
		// "delta" appears only in a secondary field here...
		{Module: ModuleAccounts, ID: 1, Title: "Acme Corp", Fields: []string{"ops@delta.example"}},
		// ...but in the title here, so this should rank first.
		{Module: ModuleAccounts, ID: 2, Title: "Delta Air"},
	}

	groups := Search("delta", candidates)

	g := findGroup(groups, ModuleAccounts)
	if g == nil || len(g.Hits) != 2 {
		t.Fatalf("hits = %+v, want both accounts", g)
	}
	if g.Hits[0].ID != 2 {
		t.Fatalf("first hit = %+v, want the title match (id 2) ranked above the field-only match", g.Hits[0])
	}
}

func TestSearchGroupsInFixedModuleOrder(t *testing.T) {
	// One matching candidate per module, supplied out of display order.
	candidates := []Candidate{
		{Module: ModuleCases, ID: 1, Title: "salt outage"},
		{Module: ModuleLeads, ID: 2, Title: "salt lead"},
		{Module: ModuleAccounts, ID: 3, Title: "salt account"},
		{Module: ModuleOpportunities, ID: 4, Title: "salt deal"},
		{Module: ModuleContacts, ID: 5, Title: "salt person"},
	}

	groups := Search("salt", candidates)

	got := make([]Module, len(groups))
	for i, g := range groups {
		got[i] = g.Module
	}
	want := []Module{ModuleAccounts, ModuleContacts, ModuleLeads, ModuleOpportunities, ModuleCases}
	if len(got) != len(want) {
		t.Fatalf("got %d groups %v, want %d in fixed order %v", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("group order = %v, want %v", got, want)
		}
	}
}

func TestSearchMatchesSecondaryFields(t *testing.T) {
	candidates := []Candidate{
		{Module: ModuleAccounts, ID: 1, Title: "Acme Corp", Fields: []string{"billing@acme.example"}},
		{Module: ModuleAccounts, ID: 2, Title: "Globex", Fields: []string{"hq@globex.example"}},
	}

	groups := Search("globex.example", candidates)

	g := findGroup(groups, ModuleAccounts)
	if g == nil || len(g.Hits) != 1 || g.Hits[0].ID != 2 {
		t.Fatalf("hits = %+v, want only Globex (matched on its email field)", g)
	}
}

func TestSearchEmptyQueryReturnsNoGroups(t *testing.T) {
	candidates := []Candidate{
		{Module: ModuleAccounts, ID: 1, Title: "Acme Corp"},
	}

	if groups := Search("", candidates); len(groups) != 0 {
		t.Fatalf("empty query returned %+v, want no groups", groups)
	}
	if groups := Search("   ", candidates); len(groups) != 0 {
		t.Fatalf("blank query returned %+v, want no groups", groups)
	}
}

func TestSearchMatchesTitleAndGroupsByModule(t *testing.T) {
	candidates := []Candidate{
		{Module: ModuleAccounts, ID: 1, Title: "Acme Corp"},
		{Module: ModuleAccounts, ID: 2, Title: "Globex"},
	}

	groups := Search("acme", candidates)

	g := findGroup(groups, ModuleAccounts)
	if g == nil {
		t.Fatalf("no accounts group in %+v", groups)
	}
	if len(g.Hits) != 1 || g.Hits[0].ID != 1 || g.Hits[0].Title != "Acme Corp" {
		t.Fatalf("accounts hits = %+v, want only Acme Corp (id 1)", g.Hits)
	}
}
