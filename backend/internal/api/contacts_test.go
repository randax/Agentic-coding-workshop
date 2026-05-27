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
	"saltcrm/internal/metadata"
)

func TestGetContactsRequiresAuth(t *testing.T) {
	_, router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/contacts", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestGetContactsReturnsVisibleRecords(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "admin@isp.example", agent.RoleAdmin, nil)
	db.Create(&contact.Contact{Name: "Ada Lovelace", Email: "ada@x.example", AccountID: 1})
	db.Create(&contact.Contact{Name: "Alan Turing", Email: "alan@x.example", AccountID: 2})

	rec := authGet(t, router, "/contacts", cookie)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got []map[string]any
	json.Unmarshal(rec.Body.Bytes(), &got)
	if len(got) != 2 {
		t.Fatalf("admin got %d contacts, want 2", len(got))
	}
}

func TestAccountContactsSubpanelReturnsOnlyThatAccountsContacts(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "admin@isp.example", agent.RoleAdmin, nil)
	acct := customer.Customer{Name: "Globex", AccountNumber: "G-1", Status: customer.StatusActive}
	db.Create(&acct)
	other := customer.Customer{Name: "Acme", AccountNumber: "A-1", Status: customer.StatusActive}
	db.Create(&other)
	db.Create(&contact.Contact{Name: "On Globex", AccountID: acct.ID})
	db.Create(&contact.Contact{Name: "Also Globex", AccountID: acct.ID})
	db.Create(&contact.Contact{Name: "On Acme", AccountID: other.ID})

	rec := authGet(t, router, "/accounts/"+strconv.FormatUint(uint64(acct.ID), 10)+"/contacts", cookie)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got []map[string]any
	json.Unmarshal(rec.Body.Bytes(), &got)
	if len(got) != 2 {
		t.Fatalf("got %d contacts for the account, want 2", len(got))
	}
}

func TestPutContactAsAgentIsForbidden(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "agent@isp.example", agent.RoleAgent, nil)
	ct := contact.Contact{Name: "Ada", AccountID: 1}
	db.Create(&ct)

	req := httptest.NewRequest(http.MethodPut, "/contacts/"+strconv.FormatUint(uint64(ct.ID), 10), nil)
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d (agents can't edit contacts)", rec.Code, http.StatusForbidden)
	}
}

func TestGetContactsMetadataReturnsContract(t *testing.T) {
	_, router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/metadata/contacts", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got metadata.ModuleMeta
	json.Unmarshal(rec.Body.Bytes(), &got)
	if got.Module != "contacts" || got.LabelSingular != "Contact" || len(got.ListView.Columns) == 0 {
		t.Errorf("meta = %+v, want contacts module with columns", got)
	}
}
