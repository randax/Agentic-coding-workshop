package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"saltcrm/internal/agent"
	"saltcrm/internal/lead"
	"saltcrm/internal/metadata"
)

func TestGetLeadsRequiresAuth(t *testing.T) {
	_, router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/leads", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestGetLeadsReturnsVisibleRecords(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "admin@isp.example", agent.RoleAdmin, nil)
	db.Create(&lead.Lead{Name: "Priya Patel", Company: "Fjord", Status: lead.StatusNew})
	db.Create(&lead.Lead{Name: "Marco Rossi", Company: "Nordlys", Status: lead.StatusWorking})

	rec := authGet(t, router, "/leads", cookie)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got []map[string]any
	json.Unmarshal(rec.Body.Bytes(), &got)
	if len(got) != 2 {
		t.Fatalf("admin got %d leads, want 2", len(got))
	}
}

func TestPutLeadAsManagerChangesStatus(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "manager@isp.example", agent.RoleManager, nil)
	l := lead.Lead{Name: "Priya Patel", Company: "Fjord", Status: lead.StatusNew}
	db.Create(&l)

	req := httptest.NewRequest(http.MethodPut, "/leads/"+strconv.FormatUint(uint64(l.ID), 10),
		strings.NewReader(`{"name":"Priya Patel","company":"Fjord","status":"qualified"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got map[string]any
	json.Unmarshal(rec.Body.Bytes(), &got)
	if got["status"] != "qualified" {
		t.Errorf("status = %v, want qualified", got["status"])
	}
}

func TestPutLeadAsAgentIsForbidden(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "agent@isp.example", agent.RoleAgent, nil)
	l := lead.Lead{Name: "Priya", Status: lead.StatusNew}
	db.Create(&l)

	req := httptest.NewRequest(http.MethodPut, "/leads/"+strconv.FormatUint(uint64(l.ID), 10), strings.NewReader(`{"name":"X","status":"new"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestGetLeadsMetadataReturnsContract(t *testing.T) {
	_, router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/metadata/leads", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var got metadata.ModuleMeta
	json.Unmarshal(rec.Body.Bytes(), &got)
	byName := map[string]metadata.Field{}
	for _, f := range got.Fields {
		byName[f.Name] = f
	}
	if got.Module != "leads" || byName["status"].Type != metadata.FieldEnum {
		t.Errorf("meta = %+v, want leads module with a status enum", got)
	}
}
