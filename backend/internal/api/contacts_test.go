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

func TestCreateContactAsAgentIsForbidden(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "agent@isp.example", agent.RoleAgent, nil)

	rec := authPost(t, router, "/contacts", cookie, `{"name":"Ada","email":"ada@x.example"}`)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d (agents can't create contacts)", rec.Code, http.StatusForbidden)
	}
}

func TestCreateContactRejectsMissingName(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "manager@isp.example", agent.RoleManager, nil)

	rec := authPost(t, router, "/contacts", cookie, `{"email":"noname@x.example"}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (name required); body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestCreateContactAssignsCreatorAsOwnerSoItIsVisible(t *testing.T) {
	db, router := newTestRouter(t)
	team := uint(1)
	cookie := loginAs(t, db, router, "manager@isp.example", agent.RoleManager, &team)
	// A contact owned by someone else on another team — must stay invisible.
	otherUser, otherTeam := uint(999), uint(2)
	db.Create(&contact.Contact{Name: "Not Mine", AssignedUserID: &otherUser, TeamID: &otherTeam})

	rec := authPost(t, router, "/contacts", cookie, `{"name":"Ada Lovelace","email":"ada@x.example","accountId":1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var created contact.Contact
	json.Unmarshal(rec.Body.Bytes(), &created)
	if created.ID == 0 || created.Name != "Ada Lovelace" {
		t.Fatalf("created = %+v, want a persisted contact with an id", created)
	}

	// The manager (a non-admin) must see the contact they just created — it is
	// only visible if the create defaulted the owner to them.
	listRec := authGet(t, router, "/contacts", cookie)
	var got []contact.Contact
	json.Unmarshal(listRec.Body.Bytes(), &got)
	if len(got) != 1 || got[0].ID != created.ID {
		t.Fatalf("manager's list = %+v, want the contact they created (owner defaulted to creator)", got)
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
