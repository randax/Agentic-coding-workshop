package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"saltcrm/internal/product"
)

func TestGetProductsReturnsCatalog(t *testing.T) {
	db, router := newTestRouter(t)
	speed := 500
	db.Create(&product.Product{
		Name: "Fiber 500", Category: product.CategoryFiber,
		MonthlyPrice: 499, Available: true, SpeedMbps: &speed,
	})
	db.Create(&product.Product{
		Name: "Mesh Pro", Category: product.CategoryRouter,
		MonthlyPrice: 99, Available: true,
	})

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got []product.Product
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v; body=%s", err, rec.Body.String())
	}
	if len(got) != 2 {
		t.Fatalf("got %d products, want 2; body=%s", len(got), rec.Body.String())
	}
}

func TestPostProductCreatesProduct(t *testing.T) {
	_, router := newTestRouter(t)
	body, _ := json.Marshal(map[string]any{
		"name":         "Fiber 1000",
		"category":     "fiber",
		"monthlyPrice": 699,
		"speedMbps":    1000,
	})

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var got product.Product
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v; body=%s", err, rec.Body.String())
	}
	if got.ID == 0 || got.Name != "Fiber 1000" || !got.Available {
		t.Errorf("created product = %+v, want id>0, name=Fiber 1000, available=true", got)
	}
}

func TestPostProductInvalidCategoryReturns400(t *testing.T) {
	_, router := newTestRouter(t)
	body, _ := json.Marshal(map[string]any{
		"name": "Mystery", "category": "satellite", "monthlyPrice": 10,
	})

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestPutProductEditsProduct(t *testing.T) {
	db, router := newTestRouter(t)
	speed := 500
	p := product.Product{
		Name: "Fiber 500", Category: product.CategoryFiber,
		MonthlyPrice: 499, Available: true, SpeedMbps: &speed,
	}
	db.Create(&p)
	body, _ := json.Marshal(map[string]any{
		"name":         "Fiber 1000",
		"category":     "fiber",
		"monthlyPrice": 699,
		"speedMbps":    1000,
	})

	req := httptest.NewRequest(http.MethodPut, "/products/"+strconv.FormatUint(uint64(p.ID), 10), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got product.Product
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v; body=%s", err, rec.Body.String())
	}
	if got.ID != p.ID || got.Name != "Fiber 1000" || got.MonthlyPrice != 699 {
		t.Errorf("updated product = %+v, want id=%d name=Fiber 1000 price=699", got, p.ID)
	}
}

func TestPutProductUnknownReturns404(t *testing.T) {
	_, router := newTestRouter(t)
	body, _ := json.Marshal(map[string]any{
		"name": "Ghost", "category": "fiber", "monthlyPrice": 1,
	})

	req := httptest.NewRequest(http.MethodPut, "/products/9999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestPutProductInvalidReturns400(t *testing.T) {
	db, router := newTestRouter(t)
	p := product.Product{Name: "Fiber 500", Category: product.CategoryFiber, MonthlyPrice: 499, Available: true}
	db.Create(&p)
	idStr := strconv.FormatUint(uint64(p.ID), 10)

	cases := map[string]map[string]any{
		"invalid category": {"name": "X", "category": "satellite", "monthlyPrice": 10},
		"negative price":   {"name": "X", "category": "fiber", "monthlyPrice": -1},
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPut, "/products/"+idStr, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
			}
		})
	}
}

func TestRetireProductMarksUnavailable(t *testing.T) {
	db, router := newTestRouter(t)
	p := product.Product{Name: "Fiber 500", Category: product.CategoryFiber, MonthlyPrice: 499, Available: true}
	db.Create(&p)

	req := httptest.NewRequest(http.MethodPost, "/products/"+strconv.FormatUint(uint64(p.ID), 10)+"/retire", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	// The catalog should now report the product as unavailable.
	listReq := httptest.NewRequest(http.MethodGet, "/products", nil)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	var got []product.Product
	json.Unmarshal(listRec.Body.Bytes(), &got)
	if len(got) != 1 || got[0].Available {
		t.Errorf("after retire, catalog = %+v, want one unavailable product", got)
	}
}

func TestUnretireProductMarksAvailable(t *testing.T) {
	db, router := newTestRouter(t)
	p := product.Product{Name: "Fiber 500", Category: product.CategoryFiber, MonthlyPrice: 499, Available: false}
	db.Create(&p)
	idStr := strconv.FormatUint(uint64(p.ID), 10)

	req := httptest.NewRequest(http.MethodPost, "/products/"+idStr+"/unretire", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	// The catalog should now report the product as available again.
	listReq := httptest.NewRequest(http.MethodGet, "/products", nil)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	var got []product.Product
	json.Unmarshal(listRec.Body.Bytes(), &got)
	if len(got) != 1 || !got[0].Available {
		t.Errorf("after unretire, catalog = %+v, want one available product", got)
	}
}

func TestUnretireUnknownProductReturns404(t *testing.T) {
	_, router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/products/9999/unretire", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestRetireUnknownProductReturns404(t *testing.T) {
	_, router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/products/9999/retire", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}
