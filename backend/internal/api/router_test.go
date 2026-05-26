package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"ispcrm/internal/customer"
	"ispcrm/internal/product"
	"ispcrm/internal/store"
	"ispcrm/internal/subscription"

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
	customers := customer.NewService(store.NewCustomerRepository(db))
	products := product.NewService(store.NewProductRepository(db))
	subscriptions := subscription.NewService(store.NewSubscriptionRepository(db), products)
	return db, NewRouter(customers, products, subscriptions)
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

func TestGetCustomersFiltersBySearchQueryParam(t *testing.T) {
	db, router := newTestRouter(t)
	db.Create(&customer.Customer{
		Name: "Ada Lovelace", Email: "ada@analytical.example", AccountNumber: "ACME-001",
		Status: customer.StatusActive, CustomerSince: time.Now(),
	})
	db.Create(&customer.Customer{
		Name: "Alan Turing", Email: "alan@bletchley.example", AccountNumber: "GLOBEX-7",
		Status: customer.StatusActive, CustomerSince: time.Now(),
	})

	req := httptest.NewRequest(http.MethodGet, "/customers?search=globex", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got []customer.Customer
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, rec.Body.String())
	}
	if len(got) != 1 || got[0].Name != "Alan Turing" {
		t.Fatalf("search=globex returned %+v, want only Alan Turing", got)
	}
}

func TestGetCustomersFiltersByStatusQueryParam(t *testing.T) {
	db, router := newTestRouter(t)
	db.Create(&customer.Customer{
		Name: "Ada Lovelace", AccountNumber: "ACME-001",
		Status: customer.StatusActive, CustomerSince: time.Now(),
	})
	db.Create(&customer.Customer{
		Name: "Alan Turing", AccountNumber: "ACME-002",
		Status: customer.StatusSuspended, CustomerSince: time.Now(),
	})

	req := httptest.NewRequest(http.MethodGet, "/customers?status=suspended", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got []customer.Customer
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, rec.Body.String())
	}
	if len(got) != 1 || got[0].Name != "Alan Turing" {
		t.Fatalf("status=suspended returned %+v, want only Alan Turing", got)
	}
}

func TestGetCustomerByIDReturnsCustomer(t *testing.T) {
	db, router := newTestRouter(t)
	c := customer.Customer{
		Name: "Grace Hopper", Email: "grace@example.com",
		AccountNumber: "ACME-003", Status: customer.StatusActive, CustomerSince: time.Now(),
	}
	db.Create(&c)

	req := httptest.NewRequest(http.MethodGet, "/customers/"+strconv.FormatUint(uint64(c.ID), 10), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got customer.Customer
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, rec.Body.String())
	}
	if got.ID != c.ID || got.Name != "Grace Hopper" || got.Email != "grace@example.com" {
		t.Errorf("got %+v, want id=%d name=Grace Hopper", got, c.ID)
	}
}

func TestGetCustomerByIDUnknownReturns404(t *testing.T) {
	_, router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/customers/9999", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestGetCustomerByIDInvalidIDReturns400(t *testing.T) {
	_, router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/customers/not-a-number", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
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
