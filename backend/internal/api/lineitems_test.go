package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"saltcrm/internal/agent"
	"saltcrm/internal/opportunity"
	"saltcrm/internal/product"
)

func TestAddLineItemSnapshotsProductAndComputesTotal(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "manager@isp.example", agent.RoleManager, nil)
	p := product.Product{Name: "Fiber 500", Category: product.CategoryFiber, MonthlyPrice: 499, Available: true}
	db.Create(&p)
	o := opportunity.Opportunity{Name: "Deal", AccountID: 1, Amount: 1000, Stage: opportunity.StageProspecting}
	db.Create(&o)

	body := `{"productId":` + strconv.FormatUint(uint64(p.ID), 10) + `,"quantity":3}`
	req := httptest.NewRequest(http.MethodPost, "/opportunities/"+strconv.FormatUint(uint64(o.ID), 10)+"/line-items", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var got map[string]any
	json.Unmarshal(rec.Body.Bytes(), &got)
	if got["productName"] != "Fiber 500" || got["lineTotal"].(float64) != 1497 {
		t.Errorf("line item = %+v, want productName Fiber 500 and lineTotal 1497", got)
	}

	// The subpanel endpoint lists it back.
	listRec := authGet(t, router, "/opportunities/"+strconv.FormatUint(uint64(o.ID), 10)+"/line-items", cookie)
	var items []map[string]any
	json.Unmarshal(listRec.Body.Bytes(), &items)
	if len(items) != 1 {
		t.Fatalf("got %d line items, want 1", len(items))
	}
}

func TestAddLineItemAsAgentIsForbidden(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "agent@isp.example", agent.RoleAgent, nil)
	o := opportunity.Opportunity{Name: "Deal", AccountID: 1, Stage: opportunity.StageProspecting}
	db.Create(&o)

	req := httptest.NewRequest(http.MethodPost, "/opportunities/"+strconv.FormatUint(uint64(o.ID), 10)+"/line-items", strings.NewReader(`{"productId":1,"quantity":1}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}
