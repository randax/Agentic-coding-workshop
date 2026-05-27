// Package search holds the cross-module global-search logic: given a query and
// a set of candidate records drawn from any module, it matches, ranks, and
// groups them by module. It is pure — it depends on no database, HTTP, or
// access model — so the ranking/grouping rules are unit-tested in isolation.
// Access scope is enforced by the caller (the HTTP handler) before candidates
// reach this package, mirroring how the dashboard composes its dashlets.
package search

import (
	"sort"
	"strings"
)

// Module identifies which CRM module a candidate/hit belongs to.
type Module string

const (
	ModuleAccounts      Module = "accounts"
	ModuleContacts      Module = "contacts"
	ModuleLeads         Module = "leads"
	ModuleOpportunities Module = "opportunities"
	ModuleCases         Module = "cases"
)

// moduleOrder fixes the display order of result groups.
var moduleOrder = []Module{
	ModuleAccounts, ModuleContacts, ModuleLeads, ModuleOpportunities, ModuleCases,
}

// Candidate is a record offered up for matching. Title is the primary label —
// it is both matched against the query and shown in results. Fields are extra
// text matched against the query but not shown (e.g. email, company, phone).
type Candidate struct {
	Module Module
	ID     uint
	Title  string
	Fields []string
}

// Relevance scores: a title match is more relevant than one found only in a
// secondary field. scoreNone means the candidate does not match at all.
const (
	scoreNone  = 0
	scoreField = 1
	scoreTitle = 2
)

// score rates how well the candidate matches the (already lower-cased) term,
// returning scoreNone when it does not match.
func (c Candidate) score(term string) int {
	if strings.Contains(strings.ToLower(c.Title), term) {
		return scoreTitle
	}
	for _, f := range c.Fields {
		if strings.Contains(strings.ToLower(f), term) {
			return scoreField
		}
	}
	return scoreNone
}

// Hit is a candidate that matched the query.
type Hit struct {
	Module Module `json:"module"`
	ID     uint   `json:"id"`
	Title  string `json:"title"`
}

// Group is the hits for a single module.
type Group struct {
	Module Module `json:"module"`
	Hits   []Hit  `json:"hits"`
}

// Search returns the candidates matching query, grouped by module.
func Search(query string, candidates []Candidate) []Group {
	term := strings.ToLower(strings.TrimSpace(query))
	if term == "" {
		return nil
	}
	// scoredHit pairs a hit with its relevance, used only for ranking here.
	type scoredHit struct {
		hit   Hit
		score int
	}
	byModule := map[Module][]scoredHit{}
	for _, c := range candidates {
		if s := c.score(term); s != scoreNone {
			byModule[c.Module] = append(byModule[c.Module],
				scoredHit{hit: Hit{Module: c.Module, ID: c.ID, Title: c.Title}, score: s})
		}
	}
	var groups []Group
	for _, m := range moduleOrder {
		scored := byModule[m]
		if len(scored) == 0 {
			continue
		}
		// Most relevant first; ties broken by ID for a stable order.
		sort.SliceStable(scored, func(i, j int) bool {
			if scored[i].score != scored[j].score {
				return scored[i].score > scored[j].score
			}
			return scored[i].hit.ID < scored[j].hit.ID
		})
		hits := make([]Hit, len(scored))
		for i, sh := range scored {
			hits[i] = sh.hit
		}
		groups = append(groups, Group{Module: m, Hits: hits})
	}
	return groups
}
