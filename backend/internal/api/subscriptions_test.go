package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"ispcrm/internal/customer"
	"ispcrm/internal/product"
	"ispcrm/internal/subscription"

	"gorm.io/gorm"
)

func seedCustomerWithSubs(t *testing.T, db *gorm.DB) (custID uint) {
	t.Helper()
	cust := customer.Customer{Name: "Ada Lovelace", AccountNumber: "A1", Status: customer.StatusActive, CustomerSince: time.Now()}
	db.Create(&cust)
	prod := product.Product{Name: "Fiber 500", Category: product.CategoryFiber, MonthlyPrice: 499, Available: true}
	db.Create(&prod)

	db.Create(&subscription.Subscription{
		CustomerID: cust.ID, ProductID: prod.ID, Status: subscription.StatusActive,
		StartDate: time.Now(), MonthlyPriceSnapshot: 499, Quantity: 1,
	})
	db.Create(&subscription.Subscription{
		CustomerID: cust.ID, ProductID: prod.ID, Status: subscription.StatusCancelled,
		StartDate: time.Now(), MonthlyPriceSnapshot: 459, Quantity: 2,
	})

	// A different customer with their own subscription, to prove filtering.
	other := customer.Customer{Name: "Bob", AccountNumber: "A2", Status: customer.StatusActive, CustomerSince: time.Now()}
	db.Create(&other)
	db.Create(&subscription.Subscription{
		CustomerID: other.ID, ProductID: prod.ID, Status: subscription.StatusActive,
		StartDate: time.Now(), MonthlyPriceSnapshot: 499, Quantity: 1,
	})
	return cust.ID
}

func getSubs(t *testing.T, router http.Handler, custID uint) (int, []subscription.Subscription) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/customers/"+strconv.FormatUint(uint64(custID), 10)+"/subscriptions", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	var subs []subscription.Subscription
	if rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &subs); err != nil {
			t.Fatalf("decode: %v; body=%s", err, rec.Body.String())
		}
	}
	return rec.Code, subs
}

func TestGetCustomerSubscriptionsReturnsThemWithProductAndSnapshot(t *testing.T) {
	db, router := newTestRouter(t)
	custID := seedCustomerWithSubs(t, db)

	code, subs := getSubs(t, router, custID)
	if code != http.StatusOK {
		t.Fatalf("status = %d, want 200", code)
	}
	if len(subs) != 2 {
		t.Fatalf("got %d subscriptions, want 2", len(subs))
	}
	for _, s := range subs {
		if s.Product.Name != "Fiber 500" {
			t.Errorf("subscription %d product name = %q, want Fiber 500", s.ID, s.Product.Name)
		}
		if s.Quantity == 0 || s.MonthlyPriceSnapshot == 0 {
			t.Errorf("subscription %d missing quantity/price snapshot: %+v", s.ID, s)
		}
	}
}

func TestGetCustomerSubscriptionsReturnsOnlyThatCustomers(t *testing.T) {
	db, router := newTestRouter(t)
	custID := seedCustomerWithSubs(t, db)

	_, subs := getSubs(t, router, custID)
	for _, s := range subs {
		if s.CustomerID != custID {
			t.Errorf("got subscription for customer %d, want only %d", s.CustomerID, custID)
		}
	}
	if len(subs) != 2 {
		t.Errorf("got %d subscriptions for the customer, want 2", len(subs))
	}
}

func TestGetCustomerSubscriptionsUnknownCustomerReturnsEmpty(t *testing.T) {
	_, router := newTestRouter(t)

	code, subs := getSubs(t, router, 9999)
	if code != http.StatusOK {
		t.Fatalf("status = %d, want 200", code)
	}
	if len(subs) != 0 {
		t.Errorf("got %d subscriptions for unknown customer, want 0", len(subs))
	}
}
