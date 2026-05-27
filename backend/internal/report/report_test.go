package report

import (
	"reflect"
	"testing"
)

func TestRunCountsRecordsPerGroup(t *testing.T) {
	records := []Record{
		{"status": "new"},
		{"status": "working"},
		{"status": "new"},
	}
	def := Definition{Module: "leads", GroupBy: "status", Aggregation: Count}

	res := Run(def, records)

	got := map[string]int{}
	for _, row := range res.Rows {
		got[row.Group] = row.Count
	}
	if got["new"] != 2 || got["working"] != 1 {
		t.Fatalf("counts = %+v, want new=2 working=1", got)
	}
}

func TestRunSumsNumericFieldPerGroup(t *testing.T) {
	records := []Record{
		{"stage": "proposal", "amount": 1000.0},
		{"stage": "proposal", "amount": 500.0},
		{"stage": "won", "amount": 9000.0},
	}
	def := Definition{Module: "opportunities", GroupBy: "stage", Aggregation: Sum, AggField: "amount"}

	res := Run(def, records)

	got := map[string]float64{}
	for _, row := range res.Rows {
		got[row.Group] = row.Value
	}
	if got["proposal"] != 1500 || got["won"] != 9000 {
		t.Fatalf("sums = %+v, want proposal=1500 won=9000", got)
	}
}

func TestRunAveragesNumericFieldPerGroup(t *testing.T) {
	records := []Record{
		{"stage": "proposal", "amount": 1000.0},
		{"stage": "proposal", "amount": 3000.0},
		{"stage": "won", "amount": 9000.0},
	}
	def := Definition{Module: "opportunities", GroupBy: "stage", Aggregation: Avg, AggField: "amount"}

	res := Run(def, records)

	got := map[string]float64{}
	for _, row := range res.Rows {
		got[row.Group] = row.Value
	}
	if got["proposal"] != 2000 || got["won"] != 9000 {
		t.Fatalf("avgs = %+v, want proposal=2000 won=9000", got)
	}
}

func TestRunFiltersWithEquals(t *testing.T) {
	records := []Record{
		{"status": "new", "source": "web"},
		{"status": "new", "source": "phone"},
		{"status": "working", "source": "web"},
	}
	def := Definition{
		Module: "leads", GroupBy: "status", Aggregation: Count,
		Filters: []Filter{{Field: "source", Operator: OpEquals, Value: "web"}},
	}

	res := Run(def, records)

	got := map[string]int{}
	for _, row := range res.Rows {
		got[row.Group] = row.Count
	}
	// Only the two web leads survive: the phone lead is excluded before grouping.
	if len(res.Rows) != 2 || got["new"] != 1 || got["working"] != 1 {
		t.Fatalf("rows = %+v, want new=1 working=1 (phone filtered out)", res.Rows)
	}
}

func TestRunFiltersWithContains(t *testing.T) {
	records := []Record{
		{"name": "Acme Corp", "stage": "won"},
		{"name": "Globex", "stage": "won"},
		{"name": "ACME Labs", "stage": "lost"},
	}
	def := Definition{
		Module: "opportunities", GroupBy: "stage", Aggregation: Count,
		Filters: []Filter{{Field: "name", Operator: OpContains, Value: "acme"}},
	}

	res := Run(def, records)

	got := map[string]int{}
	for _, row := range res.Rows {
		got[row.Group] = row.Count
	}
	// Case-insensitive substring: "Acme Corp" (won) and "ACME Labs" (lost) match;
	// "Globex" does not.
	if got["won"] != 1 || got["lost"] != 1 {
		t.Fatalf("rows = %+v, want won=1 lost=1 (case-insensitive acme match)", res.Rows)
	}
}

func TestRunFiltersWithGreaterAndLessThan(t *testing.T) {
	records := []Record{
		{"stage": "won", "amount": 500.0},
		{"stage": "won", "amount": 5000.0},
		{"stage": "won", "amount": 50000.0},
	}
	// The two range filters AND together: only the 5000 deal is > 1000 and < 10000.
	def := Definition{
		Module: "opportunities", GroupBy: "stage", Aggregation: Count,
		Filters: []Filter{
			{Field: "amount", Operator: OpGreater, Value: 1000.0},
			{Field: "amount", Operator: OpLess, Value: 10000.0},
		},
	}

	res := Run(def, records)

	if len(res.Rows) != 1 || res.Rows[0].Count != 1 {
		t.Fatalf("rows = %+v, want a single won row with count 1 (only 5000 in range)", res.Rows)
	}
}

func TestRunFilterTargetsCustomField(t *testing.T) {
	// churnRisk is a Studio custom field; on a record it is flattened to the top
	// level alongside core fields, so a filter targets it by name like any field.
	records := []Record{
		{"name": "Acme", "status": "active", "churnRisk": "high"},
		{"name": "Globex", "status": "active", "churnRisk": "low"},
		{"name": "Initech", "status": "suspended", "churnRisk": "high"},
	}
	def := Definition{
		Module: "accounts", GroupBy: "status", Aggregation: Count,
		Filters: []Filter{{Field: "churnRisk", Operator: OpEquals, Value: "high"}},
	}

	res := Run(def, records)

	got := map[string]int{}
	for _, row := range res.Rows {
		got[row.Group] = row.Count
	}
	// Only the two churnRisk=high accounts survive: one active, one suspended.
	if len(res.Rows) != 2 || got["active"] != 1 || got["suspended"] != 1 {
		t.Fatalf("rows = %+v, want active=1 suspended=1 (only churnRisk=high)", res.Rows)
	}
}

func TestRunWithoutGroupByAggregatesAll(t *testing.T) {
	records := []Record{
		{"amount": 1000.0},
		{"amount": 2000.0},
	}
	def := Definition{Module: "opportunities", Aggregation: Sum, AggField: "amount"} // no GroupBy

	res := Run(def, records)

	if len(res.Rows) != 1 || res.Rows[0].Value != 3000 {
		t.Fatalf("rows = %+v, want a single total row summing to 3000", res.Rows)
	}
	if res.Rows[0].Group != "All" {
		t.Fatalf("group = %q, want %q for an ungrouped report", res.Rows[0].Group, "All")
	}
}

func TestRunOrdersGroupsDeterministically(t *testing.T) {
	records := []Record{
		{"status": "working"},
		{"status": "new"},
		{"status": "qualified"},
	}
	def := Definition{Module: "leads", GroupBy: "status", Aggregation: Count}

	res := Run(def, records)

	var got []string
	for _, row := range res.Rows {
		got = append(got, row.Group)
	}
	want := []string{"new", "qualified", "working"} // sorted ascending by group label
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("group order = %v, want %v (sorted ascending for a stable result)", got, want)
	}
}
