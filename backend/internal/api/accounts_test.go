package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"saltcrm/internal/customer"
	"saltcrm/internal/metadata"
)

func TestGetAccountsReturnsCustomerRecords(t *testing.T) {
	db, router := newTestRouter(t)
	db.Create(&customer.Customer{
		Name: "Ada Lovelace", Email: "ada@analytical.example", AccountNumber: "ACME-001",
		Status: customer.StatusActive, CustomerSince: time.Now(),
	})
	db.Create(&customer.Customer{
		Name: "Alan Turing", Email: "alan@bletchley.example", AccountNumber: "ACME-002",
		Status: customer.StatusSuspended, CustomerSince: time.Now(),
	})

	req := httptest.NewRequest(http.MethodGet, "/accounts", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v; body=%s", err, rec.Body.String())
	}
	if len(got) != 2 {
		t.Fatalf("got %d accounts, want 2; body=%s", len(got), rec.Body.String())
	}
	names := map[string]bool{}
	for _, a := range got {
		names[a["name"].(string)] = true
	}
	if !names["Ada Lovelace"] || !names["Alan Turing"] {
		t.Errorf("account names = %v, want Ada Lovelace + Alan Turing", names)
	}
}

func TestGetAccountByIDReturnsRecord(t *testing.T) {
	db, router := newTestRouter(t)
	c := customer.Customer{
		Name: "Ada Lovelace", Email: "ada@analytical.example", AccountNumber: "ACME-001",
		Status: customer.StatusActive, CustomerSince: time.Now(),
	}
	db.Create(&c)

	req := httptest.NewRequest(http.MethodGet, "/accounts/"+strconv.FormatUint(uint64(c.ID), 10), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v; body=%s", err, rec.Body.String())
	}
	if got["name"] != "Ada Lovelace" || got["accountNumber"] != "ACME-001" {
		t.Errorf("account = %+v, want Ada Lovelace / ACME-001", got)
	}
}

func TestPutAccountUpdatesRecord(t *testing.T) {
	db, router := newTestRouter(t)
	c := customer.Customer{
		Name: "Ada Lovelace", Email: "ada@analytical.example", AccountNumber: "ACME-001",
		Status: customer.StatusActive, CustomerSince: time.Now(),
	}
	db.Create(&c)
	body := `{"name":"Ada B. Lovelace","email":"ada@analytical.example","accountNumber":"ACME-001","status":"suspended"}`

	req := httptest.NewRequest(http.MethodPut, "/accounts/"+strconv.FormatUint(uint64(c.ID), 10), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got map[string]any
	json.Unmarshal(rec.Body.Bytes(), &got)
	if got["name"] != "Ada B. Lovelace" || got["status"] != "suspended" {
		t.Errorf("updated account = %+v, want Ada B. Lovelace / suspended", got)
	}
}

func TestAccountsMetadataHasDetailEditAndSubpanels(t *testing.T) {
	_, router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/metadata/accounts", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var got metadata.ModuleMeta
	json.Unmarshal(rec.Body.Bytes(), &got)
	if len(got.DetailView.Panels) == 0 {
		t.Errorf("detailView has no panels")
	}
	if len(got.EditView.Fields) == 0 {
		t.Errorf("editView has no fields")
	}
	if len(got.Subpanels) == 0 || got.Subpanels[0].Path == "" || len(got.Subpanels[0].Columns) == 0 {
		t.Errorf("subpanels = %+v, want at least one with a path and columns", got.Subpanels)
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
	var got metadata.ModuleMeta
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v; body=%s", err, rec.Body.String())
	}
	if got.Module != "accounts" || got.LabelSingular != "Account" {
		t.Errorf("meta = %+v, want module=accounts labelSingular=Account", got)
	}
	byName := map[string]metadata.Field{}
	for _, f := range got.Fields {
		byName[f.Name] = f
	}
	if byName["status"].Type != metadata.FieldEnum || len(byName["status"].Options) != 2 {
		t.Errorf("status field = %+v, want enum with 2 options", byName["status"])
	}
	cols := got.ListView.Columns
	hasName, hasAcct := false, false
	for _, c := range cols {
		if c == "name" {
			hasName = true
		}
		if c == "accountNumber" {
			hasAcct = true
		}
	}
	if !hasName || !hasAcct {
		t.Errorf("listView columns = %v, want to include name and accountNumber", cols)
	}
}
