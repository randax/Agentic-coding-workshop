package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"saltcrm/internal/agent"
	"saltcrm/internal/customer"
	"saltcrm/internal/lead"
	"saltcrm/internal/opportunity"
)

// These tests cover the generic create endpoints added so users can make new
// records (not just edit existing ones). They assert role gating (writes are
// manager/admin, like edits) and that a created record is owned by its creator,
// so it is visible to them under own-or-team scoping.

func TestCreateLeadAsAgentIsForbidden(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "agent@isp.example", agent.RoleAgent, nil)

	rec := authPost(t, router, "/leads", cookie, `{"name":"New Lead","company":"Acme"}`)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d (agents can't create leads)", rec.Code, http.StatusForbidden)
	}
}

func TestCreateLeadAssignsCreatorAsOwnerSoItIsVisible(t *testing.T) {
	db, router := newTestRouter(t)
	team := uint(1)
	cookie := loginAs(t, db, router, "manager@isp.example", agent.RoleManager, &team)
	otherUser, otherTeam := uint(999), uint(2)
	db.Create(&lead.Lead{Name: "Not Mine", Status: lead.StatusNew, AssignedUserID: &otherUser, TeamID: &otherTeam})

	rec := authPost(t, router, "/leads", cookie, `{"name":"New Lead","company":"Acme","email":"l@x.example"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var created lead.Lead
	json.Unmarshal(rec.Body.Bytes(), &created)
	if created.ID == 0 || created.Status != lead.StatusNew {
		t.Fatalf("created = %+v, want a persisted lead defaulted to status 'new'", created)
	}

	listRec := authGet(t, router, "/leads", cookie)
	var got []lead.Lead
	json.Unmarshal(listRec.Body.Bytes(), &got)
	if len(got) != 1 || got[0].ID != created.ID {
		t.Fatalf("manager's list = %+v, want only the lead they created", got)
	}
}

func TestCreateOpportunityAsAgentIsForbidden(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "agent@isp.example", agent.RoleAgent, nil)

	rec := authPost(t, router, "/opportunities", cookie, `{"name":"Deal","accountId":1,"amount":1000,"stage":"prospecting"}`)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d (agents can't create opportunities)", rec.Code, http.StatusForbidden)
	}
}

func TestCreateOpportunityAssignsCreatorAsOwnerSoItIsVisible(t *testing.T) {
	db, router := newTestRouter(t)
	team := uint(1)
	cookie := loginAs(t, db, router, "manager@isp.example", agent.RoleManager, &team)
	otherUser, otherTeam := uint(999), uint(2)
	db.Create(&opportunity.Opportunity{Name: "Not Mine", Stage: opportunity.StageProposal, AssignedUserID: &otherUser, TeamID: &otherTeam})

	rec := authPost(t, router, "/opportunities", cookie, `{"name":"Big Deal","accountId":1,"amount":1000,"stage":"prospecting"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var created opportunity.Opportunity
	json.Unmarshal(rec.Body.Bytes(), &created)
	if created.ID == 0 || created.Probability == 0 {
		t.Fatalf("created = %+v, want a persisted opportunity with its stage probability synced", created)
	}

	listRec := authGet(t, router, "/opportunities", cookie)
	var got []opportunity.Opportunity
	json.Unmarshal(listRec.Body.Bytes(), &got)
	if len(got) != 1 || got[0].ID != created.ID {
		t.Fatalf("manager's list = %+v, want only the opportunity they created", got)
	}
}

func TestCreateAccountAsAgentIsForbidden(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "agent@isp.example", agent.RoleAgent, nil)

	rec := authPost(t, router, "/accounts", cookie, `{"name":"NewCo","email":"n@x.example","accountNumber":"NC-1","status":"active"}`)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d (agents can't create accounts)", rec.Code, http.StatusForbidden)
	}
}

func TestCreateAccountAssignsCreatorAsOwnerSoItIsVisible(t *testing.T) {
	db, router := newTestRouter(t)
	team := uint(1)
	cookie := loginAs(t, db, router, "manager@isp.example", agent.RoleManager, &team)
	otherUser, otherTeam := uint(999), uint(2)
	db.Create(&customer.Customer{Name: "Not Mine", AccountNumber: "X-1", Status: customer.StatusActive, AssignedUserID: &otherUser, TeamID: &otherTeam})

	rec := authPost(t, router, "/accounts", cookie, `{"name":"NewCo","email":"n@x.example","accountNumber":"NC-1","status":"active"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var created customer.Customer
	json.Unmarshal(rec.Body.Bytes(), &created)
	if created.ID == 0 {
		t.Fatalf("created = %+v, want a persisted account with an id", created)
	}

	listRec := authGet(t, router, "/accounts", cookie)
	var got []customer.Customer
	json.Unmarshal(listRec.Body.Bytes(), &got)
	if len(got) != 1 || got[0].ID != created.ID {
		t.Fatalf("manager's list = %+v, want only the account they created", got)
	}
}
