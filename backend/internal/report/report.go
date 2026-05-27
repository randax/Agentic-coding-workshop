// Package report holds the reporting query-builder and aggregation: given a
// report definition and a set of records drawn from a module, it filters,
// groups, and rolls them up into aggregated rows ready to chart. It is pure —
// it depends on no database, HTTP, or access model — so the query/aggregation
// rules are unit-tested in isolation. Access scope is enforced by the caller
// (the HTTP handler) before records reach this package, mirroring how the
// dashboard and global search compose their data.
package report

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Aggregation rolls a group's rows up into a single number.
type Aggregation string

const (
	// Count tallies how many records fall in each group.
	Count Aggregation = "count"
	// Sum totals the AggField across each group.
	Sum Aggregation = "sum"
	// Avg is the mean of the AggField across each group.
	Avg Aggregation = "avg"
)

// Operator compares a record's field value against a filter's value.
type Operator string

const (
	// OpEquals keeps records whose field equals the value (case-insensitive).
	OpEquals Operator = "eq"
	// OpContains keeps records whose field contains the value (case-insensitive substring).
	OpContains Operator = "contains"
	// OpGreater keeps records whose numeric field is greater than the value.
	OpGreater Operator = "gt"
	// OpLess keeps records whose numeric field is less than the value.
	OpLess Operator = "lt"
)

// Filter is one condition on a field. Field names either a core field or a
// custom field (custom values are flattened to a record's top level), so a
// filter targets custom fields exactly like core ones.
type Filter struct {
	Field    string   `json:"field"`
	Operator Operator `json:"operator"`
	Value    any      `json:"value"`
}

// Record is a single module record as a generic field bag — its JSON shape,
// with custom fields flattened to the top level, so custom fields are read by
// key exactly like core fields.
type Record map[string]any

// Definition describes a report over a module: how to partition its records
// (GroupBy) and how to roll each group up (Aggregation). An empty GroupBy
// groups every record together.
type Definition struct {
	Module      string      `json:"module"`
	Filters     []Filter    `json:"filters"`
	GroupBy     string      `json:"groupBy"`
	Aggregation Aggregation `json:"aggregation"`
	AggField    string      `json:"aggField"` // field summed/averaged; ignored for Count
}

// Row is one aggregated group: its grouping value and the rolled-up numbers.
type Row struct {
	Group string  `json:"group"`
	Count int     `json:"count"`
	Value float64 `json:"value"`
}

// Result is the report output: one aggregated row per group.
type Result struct {
	Rows []Row `json:"rows"`
}

// group accumulates the running tallies for one grouping value.
type group struct {
	count int
	sum   float64
}

// Run executes the report definition over the records, returning aggregated rows.
func Run(def Definition, records []Record) Result {
	groups := map[string]*group{}
	for _, rec := range records {
		if !matchesAll(rec, def.Filters) {
			continue
		}
		key := groupKey(rec, def.GroupBy)
		g := groups[key]
		if g == nil {
			g = &group{}
			groups[key] = g
		}
		g.count++
		g.sum += toFloat(rec[def.AggField])
	}
	var rows []Row
	for key, g := range groups {
		rows = append(rows, Row{Group: key, Count: g.count, Value: g.value(def.Aggregation)})
	}
	// Sort by group label so the result (and its chart) is stable regardless of
	// record order.
	sort.Slice(rows, func(i, j int) bool { return rows[i].Group < rows[j].Group })
	return Result{Rows: rows}
}

// value rolls the group up into the number the aggregation calls for.
func (g *group) value(agg Aggregation) float64 {
	switch agg {
	case Sum:
		return g.sum
	case Avg:
		if g.count == 0 {
			return 0
		}
		return g.sum / float64(g.count)
	default: // Count
		return float64(g.count)
	}
}

// matchesAll reports whether a record satisfies every filter (filters AND).
func matchesAll(rec Record, filters []Filter) bool {
	for _, f := range filters {
		if !f.matches(rec) {
			return false
		}
	}
	return true
}

// matches reports whether a record satisfies a single filter.
func (f Filter) matches(rec Record) bool {
	switch f.Operator {
	case OpEquals:
		return strings.EqualFold(scalarString(rec[f.Field]), scalarString(f.Value))
	case OpContains:
		return strings.Contains(
			strings.ToLower(scalarString(rec[f.Field])),
			strings.ToLower(scalarString(f.Value)),
		)
	case OpGreater:
		return toFloat(rec[f.Field]) > toFloat(f.Value)
	case OpLess:
		return toFloat(rec[f.Field]) < toFloat(f.Value)
	default:
		return true
	}
}

// groupAll is the single group's label when a report has no GroupBy.
const groupAll = "All"

// groupKey is the group a record falls into for the given GroupBy field. With no
// GroupBy, every record falls into one "All" bucket (an ungrouped total).
func groupKey(rec Record, field string) string {
	if field == "" {
		return groupAll
	}
	return scalarString(rec[field])
}

// scalarString renders a scalar field/filter value as text — its group label or
// its comparand for equality/substring filters.
// Numbers decoded from JSON arrive as float64; %v formats them without a
// trailing ".0" only via strconv, so use that for floats to keep "5" == 5.
func scalarString(v any) string {
	switch n := v.(type) {
	case nil:
		return ""
	case string:
		return n
	case float64:
		return strconv.FormatFloat(n, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(n)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// toFloat coerces a field value to a number for summing/averaging, treating
// anything non-numeric as zero. Records decoded from JSON arrive as float64;
// the integer cases cover records built directly in Go (e.g. tests).
func toFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case uint:
		return float64(n)
	default:
		return 0
	}
}
