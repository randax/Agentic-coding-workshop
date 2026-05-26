package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"ispcrm/internal/customer"
	"ispcrm/internal/store"

	"gorm.io/gorm"
)

// newTestRouter wires the real router over a fresh, throwaway SQLite database.
// This is the convention for API integration tests: exercise handlers end-to-end
// against real persistence, not mocks.
func newTestRouter(t *testing.T) (*gorm.DB, http.Handler) {
	t.Helper()
	dsn := filepath.Join(t.TempDir(), "test.db")
	db, err := store.Open(dsn)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := store.Migrate(db); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	svc := customer.NewService(store.NewCustomerRepository(db))
	return db, NewRouter(svc)
}

func TestGetCustomersReturnsSeededCustomers(t *testing.T) {
	db, router := newTestRouter(t)
	db.Create(&customer.Customer{
		Name: "Ada Lovelace", AccountNumber: "ACME-001",
		Status: customer.StatusActive, CustomerSince: time.Now(),
	})
	db.Create(&customer.Customer{
		Name: "Alan Turing", AccountNumber: "ACME-002",
		Status: customer.StatusActive, CustomerSince: time.Now(),
	})

	req := httptest.NewRequest(http.MethodGet, "/customers", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got []customer.Customer
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, rec.Body.String())
	}
	if len(got) != 2 {
		t.Fatalf("got %d customers, want 2; body=%s", len(got), rec.Body.String())
	}

	names := map[string]bool{}
	for _, c := range got {
		names[c.Name] = true
	}
	for _, want := range []string{"Ada Lovelace", "Alan Turing"} {
		if !names[want] {
			t.Errorf("expected customer %q in response", want)
		}
	}
}

func TestGetCustomersEmptyReturnsEmptyArray(t *testing.T) {
	_, router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/customers", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); body != "[]" {
		t.Errorf("empty list body = %q, want %q", body, "[]")
	}
}
