package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
