package api

import (
	"bytes"
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

// postJSON sends a JSON body to path and returns the recorder.
func postJSON(t *testing.T, router http.Handler, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func seedCustomerAndProduct(t *testing.T, db *gorm.DB, available bool) (custID, prodID uint, price float64) {
	t.Helper()
	cust := customer.Customer{Name: "Ada Lovelace", AccountNumber: "A1", Status: customer.StatusActive, CustomerSince: time.Now()}
	db.Create(&cust)
	prod := product.Product{Name: "Fiber 500", Category: product.CategoryFiber, MonthlyPrice: 499, Available: available}
	db.Create(&prod)
	return cust.ID, prod.ID, prod.MonthlyPrice
}

func TestAssignSubscriptionCreatesItWithSnapshotAndQuantity(t *testing.T) {
	db, router := newTestRouter(t)
	custID, prodID, price := seedCustomerAndProduct(t, db, true)

	rec := postJSON(t, router, "/customers/"+strconv.FormatUint(uint64(custID), 10)+"/subscriptions",
		map[string]any{"productId": prodID, "quantity": 2})

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	var created subscription.Subscription
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode: %v; body=%s", err, rec.Body.String())
	}
	if created.ID == 0 {
		t.Errorf("created subscription should have an ID")
	}
	if created.MonthlyPriceSnapshot != price {
		t.Errorf("snapshot = %v, want %v (the catalog price)", created.MonthlyPriceSnapshot, price)
	}
	if created.Status != subscription.StatusActive {
		t.Errorf("status = %q, want active", created.Status)
	}
	if created.Quantity != 2 {
		t.Errorf("quantity = %d, want 2", created.Quantity)
	}

	// It now appears in the customer's subscription list, with the product joined.
	_, subs := getSubs(t, router, custID)
	if len(subs) != 1 || subs[0].Product.Name != "Fiber 500" {
		t.Errorf("assigned subscription not listed with product: %+v", subs)
	}
}

func TestAssignRetiredProductIsRejected(t *testing.T) {
	db, router := newTestRouter(t)
	custID, prodID, _ := seedCustomerAndProduct(t, db, false)

	rec := postJSON(t, router, "/customers/"+strconv.FormatUint(uint64(custID), 10)+"/subscriptions",
		map[string]any{"productId": prodID, "quantity": 1})

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409 Conflict; body=%s", rec.Code, rec.Body.String())
	}
	_, subs := getSubs(t, router, custID)
	if len(subs) != 0 {
		t.Errorf("a rejected assignment should not be persisted, got %d", len(subs))
	}
}

func TestAssignRejectsQuantityBelowOne(t *testing.T) {
	db, router := newTestRouter(t)
	custID, prodID, _ := seedCustomerAndProduct(t, db, true)

	rec := postJSON(t, router, "/customers/"+strconv.FormatUint(uint64(custID), 10)+"/subscriptions",
		map[string]any{"productId": prodID, "quantity": 0})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}

func TestCancelSubscriptionSetsCancelledAndEndDate(t *testing.T) {
	db, router := newTestRouter(t)
	custID, prodID, _ := seedCustomerAndProduct(t, db, true)
	rec := postJSON(t, router, "/customers/"+strconv.FormatUint(uint64(custID), 10)+"/subscriptions",
		map[string]any{"productId": prodID, "quantity": 1})
	var created subscription.Subscription
	json.Unmarshal(rec.Body.Bytes(), &created)

	cancelRec := postJSON(t, router, "/subscriptions/"+strconv.FormatUint(uint64(created.ID), 10)+"/cancel", nil)
	if cancelRec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", cancelRec.Code, cancelRec.Body.String())
	}
	var cancelled subscription.Subscription
	if err := json.Unmarshal(cancelRec.Body.Bytes(), &cancelled); err != nil {
		t.Fatalf("decode: %v; body=%s", err, cancelRec.Body.String())
	}
	if cancelled.Status != subscription.StatusCancelled {
		t.Errorf("status = %q, want cancelled", cancelled.Status)
	}
	if cancelled.EndDate == nil {
		t.Errorf("cancelled subscription should have an end date")
	}
}

func TestCancelUnknownSubscriptionReturns404(t *testing.T) {
	_, router := newTestRouter(t)

	rec := postJSON(t, router, "/subscriptions/9999/cancel", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", rec.Code, rec.Body.String())
	}
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
