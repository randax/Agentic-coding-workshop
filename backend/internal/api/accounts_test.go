package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"saltcrm/internal/agent"
	"saltcrm/internal/customer"
	"saltcrm/internal/identity"

	"gorm.io/gorm"
)

// loginAs creates a user with the given role/team and returns a session cookie
// for them by logging in through the real endpoint.
func loginAs(t *testing.T, db *gorm.DB, router http.Handler, email string, role agent.Role, teamID *uint) string {
	t.Helper()
	hash, _ := identity.HashPassword("pw")
	db.Create(&agent.Agent{Name: email, Email: email, PasswordHash: hash, Role: role, TeamID: teamID})

	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"email":"`+email+`","password":"pw"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("loginAs %s: status %d; body=%s", email, rec.Code, rec.Body.String())
	}
	return rec.Header().Get("Set-Cookie")
}

func authGet(t *testing.T, router http.Handler, path, cookie string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestGetAccountsRequiresAuth(t *testing.T) {
	_, router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/accounts", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d (auth required); body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestGetAccountsReturnsVisibleRecords(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "admin@isp.example", agent.RoleAdmin, nil)
	db.Create(&customer.Customer{Name: "Ada Lovelace", AccountNumber: "ACME-001", Status: customer.StatusActive, CustomerSince: time.Now()})
	db.Create(&customer.Customer{Name: "Alan Turing", AccountNumber: "ACME-002", Status: customer.StatusActive, CustomerSince: time.Now()})

	rec := authGet(t, router, "/accounts", cookie)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got []map[string]any
	json.Unmarshal(rec.Body.Bytes(), &got)
	if len(got) != 2 { // admin sees all
		t.Fatalf("admin got %d accounts, want 2", len(got))
	}
}

func TestAccountsListScopedToOwnOrTeam(t *testing.T) {
	db, router := newTestRouter(t)
	teamA := uint(1)
	teamB := uint(2)
	// Log in as a plain agent on team A.
	cookie := loginAs(t, db, router, "agent@isp.example", agent.RoleAgent, &teamA)
	var me agent.Agent
	db.Where("email = ?", "agent@isp.example").First(&me)

	mine := customer.Customer{Name: "Mine", AccountNumber: "A-1", Status: customer.StatusActive, CustomerSince: time.Now(), AssignedUserID: &me.ID}
	db.Create(&mine)
	teammate := customer.Customer{Name: "Teammate", AccountNumber: "A-2", Status: customer.StatusActive, CustomerSince: time.Now(), TeamID: &teamA}
	db.Create(&teammate)
	other := customer.Customer{Name: "OtherTeam", AccountNumber: "A-3", Status: customer.StatusActive, CustomerSince: time.Now(), TeamID: &teamB}
	db.Create(&other)

	rec := authGet(t, router, "/accounts", cookie)
	var got []map[string]any
	json.Unmarshal(rec.Body.Bytes(), &got)

	names := map[string]bool{}
	for _, a := range got {
		names[a["name"].(string)] = true
	}
	if !names["Mine"] || !names["Teammate"] {
		t.Errorf("agent should see own + team records, got %v", names)
	}
	if names["OtherTeam"] {
		t.Errorf("agent should NOT see another team's record, got %v", names)
	}
}

func TestGetAccountByIDReturnsRecord(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "admin@isp.example", agent.RoleAdmin, nil)
	c := customer.Customer{Name: "Ada Lovelace", AccountNumber: "ACME-001", Status: customer.StatusActive, CustomerSince: time.Now()}
	db.Create(&c)

	rec := authGet(t, router, "/accounts/"+strconv.FormatUint(uint64(c.ID), 10), cookie)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got map[string]any
	json.Unmarshal(rec.Body.Bytes(), &got)
	if got["name"] != "Ada Lovelace" {
		t.Errorf("account = %+v, want Ada Lovelace", got)
	}
}

func putAccount(t *testing.T, router http.Handler, id uint, cookie, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, "/accounts/"+strconv.FormatUint(uint64(id), 10), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestPutAccountAsManagerUpdatesRecord(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "manager@isp.example", agent.RoleManager, nil)
	c := customer.Customer{Name: "Ada Lovelace", Email: "ada@x.example", AccountNumber: "ACME-001", Status: customer.StatusActive, CustomerSince: time.Now()}
	db.Create(&c)

	rec := putAccount(t, router, c.ID, cookie, `{"name":"Ada B. Lovelace","email":"ada@x.example","accountNumber":"ACME-001","status":"suspended"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got map[string]any
	json.Unmarshal(rec.Body.Bytes(), &got)
	if got["name"] != "Ada B. Lovelace" || got["status"] != "suspended" {
		t.Errorf("updated = %+v, want Ada B. Lovelace / suspended", got)
	}
}

func TestPutAccountAsAgentIsForbidden(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "agent@isp.example", agent.RoleAgent, nil)
	c := customer.Customer{Name: "Ada Lovelace", Email: "ada@x.example", AccountNumber: "ACME-001", Status: customer.StatusActive, CustomerSince: time.Now()}
	db.Create(&c)

	rec := putAccount(t, router, c.ID, cookie, `{"name":"Nope","email":"ada@x.example","accountNumber":"ACME-001","status":"active"}`)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d (agents can't edit accounts); body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestGetAccountsMetadataReturnsContract(t *testing.T) {
	_, router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/metadata/accounts", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}
