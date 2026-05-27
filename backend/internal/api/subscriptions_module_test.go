package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"saltcrm/internal/agent"
	"saltcrm/internal/customer"
	"saltcrm/internal/metadata"
	"saltcrm/internal/product"
	"saltcrm/internal/subscription"
)

func TestGetSubscriptionsRequiresAuth(t *testing.T) {
	_, router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/subscriptions", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestGetSubscriptionsListsAllWithProduct(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "admin@isp.example", agent.RoleAdmin, nil)
	cust := customer.Customer{Name: "Ada", AccountNumber: "A-1", Status: customer.StatusActive}
	db.Create(&cust)
	p := product.Product{Name: "Fiber 500", Category: product.CategoryFiber, MonthlyPrice: 499, Available: true}
	db.Create(&p)
	db.Create(&subscription.Subscription{CustomerID: cust.ID, ProductID: p.ID, Status: subscription.StatusActive, MonthlyPriceSnapshot: 499, Quantity: 1})

	rec := authGet(t, router, "/subscriptions", cookie)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got []subscription.Subscription
	json.Unmarshal(rec.Body.Bytes(), &got)
	if len(got) != 1 || got[0].Product.Name != "Fiber 500" {
		t.Fatalf("subscriptions = %+v, want 1 with its product preloaded", got)
	}
}

func TestAccountsMetadataHasSubscriptionsSubpanel(t *testing.T) {
	_, router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/metadata/accounts", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var got metadata.ModuleMeta
	json.Unmarshal(rec.Body.Bytes(), &got)
	labels := map[string]bool{}
	for _, sp := range got.Subpanels {
		labels[sp.Label] = true
	}
	if !labels["Cases"] || !labels["Subscriptions"] || !labels["Contacts"] {
		t.Errorf("account subpanels = %v, want Contacts + Cases + Subscriptions", labels)
	}
}
