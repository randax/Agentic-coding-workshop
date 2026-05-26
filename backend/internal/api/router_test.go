package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"ispcrm/internal/agent"
	"ispcrm/internal/customer"
	"ispcrm/internal/product"
	"ispcrm/internal/store"
	"ispcrm/internal/subscription"
	"ispcrm/internal/supportcase"

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
	agents := agent.NewService(store.NewAgentRepository(db))
	cases := supportcase.NewService(store.NewCaseRepository(db))
	return db, NewRouter(customers, products, subscriptions, agents, cases)
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

func TestGetAgentsReturnsAgents(t *testing.T) {
	db, router := newTestRouter(t)
	db.Create(&agent.Agent{Name: "Sam Carter", Email: "sam@isp.example"})
	db.Create(&agent.Agent{Name: "Robin Diaz", Email: "robin@isp.example"})

	req := httptest.NewRequest(http.MethodGet, "/agents", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got []agent.Agent
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, rec.Body.String())
	}
	if len(got) != 2 {
		t.Fatalf("got %d agents, want 2; body=%s", len(got), rec.Body.String())
	}
}

func TestPostCustomerCaseCreatesCase(t *testing.T) {
	db, router := newTestRouter(t)
	cust := customer.Customer{Name: "Ada Lovelace", AccountNumber: "ISP-1", Status: customer.StatusActive, CustomerSince: time.Now()}
	db.Create(&cust)
	body := `{"subject":"No internet","description":"Down since 8am","category":"connectivity","priority":"high"}`

	req := httptest.NewRequest(http.MethodPost, "/customers/"+strconv.FormatUint(uint64(cust.ID), 10)+"/cases", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var got supportcase.Case
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, rec.Body.String())
	}
	if got.ID == 0 || got.CustomerID != cust.ID {
		t.Errorf("case = %+v, want assigned ID and customer %d", got, cust.ID)
	}
	if got.Subject != "No internet" || got.Category != supportcase.CategoryConnectivity || got.Priority != supportcase.PriorityHigh {
		t.Errorf("case = %+v, want subject/category/priority from body", got)
	}
	if got.Status != supportcase.StatusOpen {
		t.Errorf("new case status = %q, want %q", got.Status, supportcase.StatusOpen)
	}
}

func TestPostCustomerCaseInvalidCategoryReturns400(t *testing.T) {
	db, router := newTestRouter(t)
	cust := customer.Customer{Name: "Ada Lovelace", AccountNumber: "ISP-1", Status: customer.StatusActive, CustomerSince: time.Now()}
	db.Create(&cust)
	body := `{"subject":"Help","category":"nonsense","priority":"high"}`

	req := httptest.NewRequest(http.MethodPost, "/customers/"+strconv.FormatUint(uint64(cust.ID), 10)+"/cases", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestGetCaseReturnsCaseWithCommentsInChronologicalOrder(t *testing.T) {
	db, router := newTestRouter(t)
	cust := customer.Customer{Name: "Ada Lovelace", AccountNumber: "ISP-1", Status: customer.StatusActive, CustomerSince: time.Now()}
	db.Create(&cust)
	ag := agent.Agent{Name: "Sam Carter", Email: "sam@isp.example"}
	db.Create(&ag)
	kase := supportcase.Case{
		CustomerID: cust.ID, Subject: "No internet", Description: "Down since 8am",
		Category: supportcase.CategoryConnectivity, Priority: supportcase.PriorityHigh,
		Status: supportcase.StatusOpen, AssignedAgentID: &ag.ID,
	}
	db.Omit("AssignedAgent").Create(&kase)

	// Insert comments out of chronological order; the response must be sorted.
	base := time.Now()
	db.Create(&supportcase.CaseComment{CaseID: kase.ID, Body: "Third", AuthorAgentID: &ag.ID, CreatedAt: base.Add(2 * time.Hour)})
	db.Create(&supportcase.CaseComment{CaseID: kase.ID, Body: "First", AuthorAgentID: &ag.ID, CreatedAt: base})
	db.Create(&supportcase.CaseComment{CaseID: kase.ID, Body: "Second", AuthorAgentID: &ag.ID, CreatedAt: base.Add(1 * time.Hour)})

	req := httptest.NewRequest(http.MethodGet, "/cases/"+strconv.FormatUint(uint64(kase.ID), 10), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got supportcase.Case
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, rec.Body.String())
	}
	if got.Subject != "No internet" || got.Description != "Down since 8am" {
		t.Errorf("case = %+v, want subject/description set", got)
	}
	if got.AssignedAgent == nil || got.AssignedAgent.Name != "Sam Carter" {
		t.Errorf("assigned agent not populated: %+v", got.AssignedAgent)
	}
	if len(got.Comments) != 3 {
		t.Fatalf("got %d comments, want 3; body=%s", len(got.Comments), rec.Body.String())
	}
	wantOrder := []string{"First", "Second", "Third"}
	for i, want := range wantOrder {
		if got.Comments[i].Body != want {
			t.Errorf("comment[%d].Body = %q, want %q (timeline must be chronological)", i, got.Comments[i].Body, want)
		}
	}
	if got.Comments[0].AuthorAgent == nil || got.Comments[0].AuthorAgent.Name != "Sam Carter" {
		t.Errorf("comment author not populated: %+v", got.Comments[0].AuthorAgent)
	}
}

func TestGetCaseUnknownReturns404(t *testing.T) {
	_, router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/cases/9999", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestGetCustomerCasesReturnsCasesWithAssignedAgent(t *testing.T) {
	db, router := newTestRouter(t)
	cust := customer.Customer{Name: "Ada Lovelace", AccountNumber: "ISP-1", Status: customer.StatusActive, CustomerSince: time.Now()}
	db.Create(&cust)
	other := customer.Customer{Name: "Alan Turing", AccountNumber: "ISP-2", Status: customer.StatusActive, CustomerSince: time.Now()}
	db.Create(&other)
	ag := agent.Agent{Name: "Sam Carter", Email: "sam@isp.example"}
	db.Create(&ag)

	db.Omit("AssignedAgent").Create(&supportcase.Case{
		CustomerID: cust.ID, Subject: "No internet", Category: supportcase.CategoryConnectivity,
		Priority: supportcase.PriorityHigh, Status: supportcase.StatusOpen, AssignedAgentID: &ag.ID,
	})
	db.Omit("AssignedAgent").Create(&supportcase.Case{
		CustomerID: other.ID, Subject: "Other customer's case", Category: supportcase.CategoryBilling,
		Priority: supportcase.PriorityLow, Status: supportcase.StatusOpen,
	})

	req := httptest.NewRequest(http.MethodGet, "/customers/"+strconv.FormatUint(uint64(cust.ID), 10)+"/cases", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got []supportcase.Case
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, rec.Body.String())
	}
	if len(got) != 1 {
		t.Fatalf("got %d cases, want 1 (only this customer's); body=%s", len(got), rec.Body.String())
	}
	if got[0].Subject != "No internet" || got[0].Priority != supportcase.PriorityHigh {
		t.Errorf("case = %+v, want subject 'No internet' priority high", got[0])
	}
	if got[0].AssignedAgent == nil || got[0].AssignedAgent.Name != "Sam Carter" {
		t.Errorf("assigned agent not populated: %+v", got[0].AssignedAgent)
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

func TestPostCustomerCreatesCustomer(t *testing.T) {
	_, router := newTestRouter(t)
	body := `{"name":"Grace Hopper","email":"grace@navy.example","phone":"555-0100","serviceAddress":"1 Navy Yard","accountNumber":"ACME-003","status":"active"}`

	req := httptest.NewRequest(http.MethodPost, "/customers", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var got customer.Customer
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, rec.Body.String())
	}
	if got.ID == 0 || got.Name != "Grace Hopper" || got.AccountNumber != "ACME-003" {
		t.Errorf("created customer = %+v, want assigned ID and Grace Hopper", got)
	}
}

func TestPostCustomerInvalidReturns400(t *testing.T) {
	_, router := newTestRouter(t)
	body := `{"email":"x@example.com","accountNumber":"ACME-9","status":"active"}` // missing name

	req := httptest.NewRequest(http.MethodPost, "/customers", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestPutCustomerUpdatesCustomer(t *testing.T) {
	db, router := newTestRouter(t)
	c := customer.Customer{
		Name: "Grace Hopper", Email: "grace@navy.example", AccountNumber: "ACME-9",
		Status: customer.StatusActive, CustomerSince: time.Now(),
	}
	db.Create(&c)
	body := `{"name":"Grace B. Hopper","email":"ghopper@navy.example","phone":"555-0100","serviceAddress":"1 Navy Yard","accountNumber":"ACME-9","status":"suspended"}`

	req := httptest.NewRequest(http.MethodPut, "/customers/"+strconv.FormatUint(uint64(c.ID), 10), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got customer.Customer
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, rec.Body.String())
	}
	if got.ID != c.ID || got.Name != "Grace B. Hopper" || got.Status != customer.StatusSuspended {
		t.Errorf("updated customer = %+v, want id=%d Grace B. Hopper suspended", got, c.ID)
	}
}

func TestPutCustomerUnknownReturns404(t *testing.T) {
	_, router := newTestRouter(t)
	body := `{"name":"Nobody","email":"n@example.com","accountNumber":"X","status":"active"}`

	req := httptest.NewRequest(http.MethodPut, "/customers/9999", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
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
