package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"saltcrm/internal/agent"
	"saltcrm/internal/contact"
	"saltcrm/internal/customer"
	"saltcrm/internal/lead"
	"saltcrm/internal/opportunity"
)

func postConvert(t *testing.T, router http.Handler, leadID uint, cookie string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/leads/"+strconv.FormatUint(uint64(leadID), 10)+"/convert", nil)
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestConvertLeadCreatesAccountContactOpportunityAndMarksLead(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "agent@isp.example", agent.RoleAgent, nil)
	l := lead.Lead{Name: "Priya Patel", Company: "Fjord Logistics", Email: "priya@fjord.example", Status: lead.StatusQualified}
	db.Create(&l)

	rec := postConvert(t, router, l.ID, cookie)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var res struct {
		AccountID     uint `json:"accountId"`
		ContactID     uint `json:"contactId"`
		OpportunityID uint `json:"opportunityId"`
	}
	json.Unmarshal(rec.Body.Bytes(), &res)
	if res.AccountID == 0 || res.ContactID == 0 || res.OpportunityID == 0 {
		t.Fatalf("result = %+v, want all three created", res)
	}

	// The records exist and are wired together.
	var acct customer.Customer
	db.First(&acct, res.AccountID)
	if acct.Name != "Fjord Logistics" {
		t.Errorf("account name = %q, want Fjord Logistics", acct.Name)
	}
	var ct contact.Contact
	db.First(&ct, res.ContactID)
	if ct.Name != "Priya Patel" || ct.AccountID != res.AccountID {
		t.Errorf("contact = %+v, want Priya Patel on the new account", ct)
	}
	var opp opportunity.Opportunity
	db.First(&opp, res.OpportunityID)
	if opp.AccountID != res.AccountID {
		t.Errorf("opportunity not linked to the new account: %+v", opp)
	}

	// The lead is marked converted and linked.
	var reloaded lead.Lead
	db.First(&reloaded, l.ID)
	if reloaded.Status != lead.StatusConverted ||
		reloaded.ConvertedAccountID == nil || *reloaded.ConvertedAccountID != res.AccountID {
		t.Errorf("lead after convert = %+v, want converted + linked to account %d", reloaded, res.AccountID)
	}
}

func TestConvertAlreadyConvertedReturns409(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "agent@isp.example", agent.RoleAgent, nil)
	l := lead.Lead{Name: "Priya", Status: lead.StatusConverted}
	db.Create(&l)

	rec := postConvert(t, router, l.ID, cookie)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestConvertRequiresAuth(t *testing.T) {
	_, router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/leads/1/convert", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
