package api

import (
	"net/http"

	"saltcrm/internal/access"
	"saltcrm/internal/contact"
	"saltcrm/internal/customer"
	"saltcrm/internal/lead"
	"saltcrm/internal/opportunity"
	"saltcrm/internal/search"
	"saltcrm/internal/supportcase"

	"github.com/gin-gonic/gin"
)

// searchHandler powers the global search box: it gathers candidate records from
// every searchable module — restricted to what the signed-in user may see — and
// delegates the matching/ranking/grouping to the pure search package. It holds
// no business logic of its own, composing existing domain services like the
// dashboard handler does.
type searchHandler struct {
	customers     *customer.Service
	contacts      *contact.Service
	leads         *lead.Service
	opportunities *opportunity.Service
	cases         *supportcase.Service
}

func (h *searchHandler) get(c *gin.Context) {
	user, ok := currentUser(c) // route is behind requireAuth, so this is set
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	ctx := c.Request.Context()
	fail := func() { c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"}) }

	var candidates []search.Candidate

	// Accounts (customers): own-or-team visible only.
	accounts, err := h.customers.List(ctx, customer.Filter{})
	if err != nil {
		fail()
		return
	}
	var visibleAccountIDs []uint
	for _, a := range accounts {
		if access.Visible(user, a.AssignedUserID, a.TeamID) {
			visibleAccountIDs = append(visibleAccountIDs, a.ID)
			candidates = append(candidates, search.Candidate{
				Module: search.ModuleAccounts, ID: a.ID, Title: a.Name,
				Fields: []string{a.Email, a.AccountNumber, a.Phone},
			})
		}
	}

	// Contacts.
	contacts, err := h.contacts.List(ctx)
	if err != nil {
		fail()
		return
	}
	for _, ct := range contacts {
		if access.Visible(user, ct.AssignedUserID, ct.TeamID) {
			candidates = append(candidates, search.Candidate{
				Module: search.ModuleContacts, ID: ct.ID, Title: ct.Name,
				Fields: []string{ct.Email, ct.Phone, ct.Title},
			})
		}
	}

	// Leads.
	leads, err := h.leads.List(ctx)
	if err != nil {
		fail()
		return
	}
	for _, l := range leads {
		if access.Visible(user, l.AssignedUserID, l.TeamID) {
			candidates = append(candidates, search.Candidate{
				Module: search.ModuleLeads, ID: l.ID, Title: l.Name,
				Fields: []string{l.Email, l.Company, l.Phone},
			})
		}
	}

	// Opportunities.
	opps, err := h.opportunities.List(ctx)
	if err != nil {
		fail()
		return
	}
	for _, o := range opps {
		if access.Visible(user, o.AssignedUserID, o.TeamID) {
			candidates = append(candidates, search.Candidate{
				Module: search.ModuleOpportunities, ID: o.ID, Title: o.Name,
			})
		}
	}

	// Cases have no owner/team of their own; they inherit the visibility of
	// their parent account. Gathering them per visible account therefore scopes
	// them correctly without any case-level access fields.
	for _, accountID := range visibleAccountIDs {
		cases, err := h.cases.ListForCustomer(ctx, accountID)
		if err != nil {
			fail()
			return
		}
		for _, k := range cases {
			candidates = append(candidates, search.Candidate{
				Module: search.ModuleCases, ID: k.ID, Title: k.Subject,
				Fields: []string{k.Description},
			})
		}
	}

	groups := search.Search(c.Query("q"), candidates)
	if groups == nil {
		groups = []search.Group{}
	}
	c.JSON(http.StatusOK, gin.H{"groups": groups})
}
