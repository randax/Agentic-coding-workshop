package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"saltcrm/internal/agent"
	"saltcrm/internal/contact"
	"saltcrm/internal/customer"
	"saltcrm/internal/lead"
	"saltcrm/internal/opportunity"
	"saltcrm/internal/search"
	"saltcrm/internal/supportcase"
)

// searchGroups runs an authenticated /search query and decodes its groups.
func searchGroups(t *testing.T, router http.Handler, query, cookie string) []search.Group {
	t.Helper()
	rec := authGet(t, router, "/search?q="+query, cookie)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var body struct {
		Groups []search.Group `json:"groups"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode search: %v; body=%s", err, rec.Body.String())
	}
	return body.Groups
}

// hitInGroups reports whether any group for module m contains a hit with id.
func hitInGroups(groups []search.Group, m search.Module, id uint) bool {
	for _, g := range groups {
		if g.Module != m {
			continue
		}
		for _, h := range g.Hits {
			if h.ID == id {
				return true
			}
		}
	}
	return false
}

func TestSearchAcrossModulesGroupsHits(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "admin@isp.example", agent.RoleAdmin, nil)

	acct := customer.Customer{Name: "Northwind Traders", AccountNumber: "NW-1", Status: customer.StatusActive, CustomerSince: time.Now()}
	db.Create(&acct)
	con := contact.Contact{Name: "Nina North", Email: "nina@nw.example", AccountID: acct.ID}
	db.Create(&con)
	ld := lead.Lead{Name: "Northstar Co", Status: lead.StatusNew}
	db.Create(&ld)
	opp := opportunity.Opportunity{Name: "North expansion deal", Stage: opportunity.StageProspecting, AccountID: acct.ID}
	db.Create(&opp)
	// A record in a module that does not match the query must not appear.
	other := lead.Lead{Name: "Southgate Ltd", Status: lead.StatusNew}
	db.Create(&other)

	groups := searchGroups(t, router, "north", cookie)

	if !hitInGroups(groups, search.ModuleAccounts, acct.ID) {
		t.Errorf("accounts group missing Northwind Traders; groups=%+v", groups)
	}
	if !hitInGroups(groups, search.ModuleContacts, con.ID) {
		t.Errorf("contacts group missing Nina North; groups=%+v", groups)
	}
	if !hitInGroups(groups, search.ModuleLeads, ld.ID) {
		t.Errorf("leads group missing Northstar Co; groups=%+v", groups)
	}
	if !hitInGroups(groups, search.ModuleOpportunities, opp.ID) {
		t.Errorf("opportunities group missing North expansion deal; groups=%+v", groups)
	}
	if hitInGroups(groups, search.ModuleLeads, other.ID) {
		t.Errorf("non-matching lead Southgate Ltd leaked into results; groups=%+v", groups)
	}
}

func TestSearchRespectsAccessScope(t *testing.T) {
	db, router := newTestRouter(t)
	myTeam := uint(10)
	cookie := loginAs(t, db, router, "agent@isp.example", agent.RoleAgent, &myTeam)
	otherTeam := uint(20)

	mine := customer.Customer{Name: "Acme Mine", AccountNumber: "A-1", Status: customer.StatusActive, CustomerSince: time.Now(), TeamID: &myTeam}
	db.Create(&mine)
	theirs := customer.Customer{Name: "Acme Theirs", AccountNumber: "A-2", Status: customer.StatusActive, CustomerSince: time.Now(), TeamID: &otherTeam}
	db.Create(&theirs)

	groups := searchGroups(t, router, "acme", cookie)

	if !hitInGroups(groups, search.ModuleAccounts, mine.ID) {
		t.Errorf("own-team account missing from results; groups=%+v", groups)
	}
	if hitInGroups(groups, search.ModuleAccounts, theirs.ID) {
		t.Errorf("leaked an account outside the user's team; groups=%+v", groups)
	}
}

func TestSearchCasesInheritAccountVisibility(t *testing.T) {
	db, router := newTestRouter(t)
	myTeam := uint(10)
	cookie := loginAs(t, db, router, "agent@isp.example", agent.RoleAgent, &myTeam)
	otherTeam := uint(20)

	visibleAcct := customer.Customer{Name: "Visible Co", AccountNumber: "V-1", Status: customer.StatusActive, CustomerSince: time.Now(), TeamID: &myTeam}
	db.Create(&visibleAcct)
	hiddenAcct := customer.Customer{Name: "Hidden Co", AccountNumber: "H-1", Status: customer.StatusActive, CustomerSince: time.Now(), TeamID: &otherTeam}
	db.Create(&hiddenAcct)

	mk := func(accountID uint, subject string) uint {
		k := supportcase.Case{
			CustomerID: accountID, Subject: subject, Category: supportcase.CategoryGeneral,
			Priority: supportcase.PriorityLow, Status: supportcase.StatusOpen,
		}
		db.Omit("AssignedAgent").Create(&k)
		return k.ID
	}
	visibleCase := mk(visibleAcct.ID, "Fiber outage downtown")
	hiddenCase := mk(hiddenAcct.ID, "Fiber outage uptown")

	groups := searchGroups(t, router, "outage", cookie)

	if !hitInGroups(groups, search.ModuleCases, visibleCase) {
		t.Errorf("case under a visible account missing; groups=%+v", groups)
	}
	if hitInGroups(groups, search.ModuleCases, hiddenCase) {
		t.Errorf("leaked a case under an account outside the user's scope; groups=%+v", groups)
	}
}

func TestSearchRequiresAuth(t *testing.T) {
	_, router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/search?q=acme", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d (auth required)", rec.Code, http.StatusUnauthorized)
	}
}
