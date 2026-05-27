package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"saltcrm/internal/agent"
	"saltcrm/internal/customer"
	"saltcrm/internal/lead"
	"saltcrm/internal/opportunity"
	"saltcrm/internal/report"
)

// authPost sends an authenticated JSON POST (raw body) and returns the recorder.
func authPost(t *testing.T, router http.Handler, path, cookie, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// rowsByGroup indexes a report result by group label for easy assertions.
func rowsByGroup(res report.Result) map[string]report.Row {
	out := map[string]report.Row{}
	for _, row := range res.Rows {
		out[row.Group] = row
	}
	return out
}

func TestRunReportRequiresAuth(t *testing.T) {
	_, router := newTestRouter(t)

	rec := authPost(t, router, "/reports/run", "", `{"module":"leads","groupBy":"status","aggregation":"count"}`)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d (auth required); body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestRunReportCountsLeadsByStatus(t *testing.T) {
	db, router := newTestRouter(t)
	admin := loginAs(t, db, router, "admin@isp.example", agent.RoleAdmin, nil)
	db.Create(&lead.Lead{Name: "A", Status: lead.StatusNew})
	db.Create(&lead.Lead{Name: "B", Status: lead.StatusNew})
	db.Create(&lead.Lead{Name: "C", Status: lead.StatusWorking})

	rec := authPost(t, router, "/reports/run", admin, `{"module":"leads","groupBy":"status","aggregation":"count"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var res report.Result
	json.Unmarshal(rec.Body.Bytes(), &res)
	got := rowsByGroup(res)
	if got["new"].Count != 2 || got["working"].Count != 1 {
		t.Fatalf("rows = %+v, want new=2 working=1", res.Rows)
	}
}

func TestRunReportSumsOpportunityAmountByStage(t *testing.T) {
	db, router := newTestRouter(t)
	admin := loginAs(t, db, router, "admin@isp.example", agent.RoleAdmin, nil)
	db.Create(&opportunity.Opportunity{Name: "D1", Stage: opportunity.StageProposal, Amount: 1000})
	db.Create(&opportunity.Opportunity{Name: "D2", Stage: opportunity.StageProposal, Amount: 500})
	db.Create(&opportunity.Opportunity{Name: "D3", Stage: opportunity.StageClosedWon, Amount: 9000})

	rec := authPost(t, router, "/reports/run", admin,
		`{"module":"opportunities","groupBy":"stage","aggregation":"sum","aggField":"amount"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var res report.Result
	json.Unmarshal(rec.Body.Bytes(), &res)
	got := rowsByGroup(res)
	if got["proposal"].Value != 1500 || got["closed_won"].Value != 9000 {
		t.Fatalf("rows = %+v, want proposal=1500 closed_won=9000", res.Rows)
	}
}

func TestRunReportFiltersOnCustomField(t *testing.T) {
	db, router := newTestRouter(t)
	admin := loginAs(t, db, router, "admin@isp.example", agent.RoleAdmin, nil)

	// Admin defines a custom field on accounts via Studio.
	if rec := addField(t, router, admin, `{"module":"accounts","name":"churnRisk","type":"enum","label":"Churn risk","options":["low","high"]}`); rec.Code != http.StatusCreated {
		t.Fatalf("addField status = %d; body=%s", rec.Code, rec.Body.String())
	}
	// Two active accounts differing only in the custom field's value.
	db.Create(&customer.Customer{Name: "Acme", AccountNumber: "A1", Status: customer.StatusActive, CustomFields: map[string]any{"churnRisk": "high"}})
	db.Create(&customer.Customer{Name: "Globex", AccountNumber: "A2", Status: customer.StatusActive, CustomFields: map[string]any{"churnRisk": "low"}})

	rec := authPost(t, router, "/reports/run", admin,
		`{"module":"accounts","groupBy":"status","aggregation":"count","filters":[{"field":"churnRisk","operator":"eq","value":"high"}]}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var res report.Result
	json.Unmarshal(rec.Body.Bytes(), &res)
	got := rowsByGroup(res)
	// Only the churnRisk=high account is counted.
	if len(res.Rows) != 1 || got["active"].Count != 1 {
		t.Fatalf("rows = %+v, want a single active row with count 1 (custom-field filter)", res.Rows)
	}
}

func TestSaveListAndRunSavedReport(t *testing.T) {
	db, router := newTestRouter(t)
	admin := loginAs(t, db, router, "admin@isp.example", agent.RoleAdmin, nil)
	db.Create(&lead.Lead{Name: "A", Status: lead.StatusNew})
	db.Create(&lead.Lead{Name: "B", Status: lead.StatusWorking})

	// Save a report.
	saveRec := authPost(t, router, "/reports", admin,
		`{"name":"Leads by status","definition":{"module":"leads","groupBy":"status","aggregation":"count"}}`)
	if saveRec.Code != http.StatusCreated {
		t.Fatalf("save status = %d, want 201; body=%s", saveRec.Code, saveRec.Body.String())
	}
	var saved report.Saved
	json.Unmarshal(saveRec.Body.Bytes(), &saved)
	if saved.ID == 0 {
		t.Fatalf("saved report = %+v, want an assigned ID", saved)
	}

	// It appears in the saved-report list.
	listRec := authGet(t, router, "/reports", admin)
	var list []report.Saved
	json.Unmarshal(listRec.Body.Bytes(), &list)
	if len(list) != 1 || list[0].Name != "Leads by status" {
		t.Fatalf("list = %+v, want the one saved report", list)
	}

	// Re-running it aggregates the current data.
	runRec := authPost(t, router, "/reports/"+strconv.FormatUint(uint64(saved.ID), 10)+"/run", admin, "")
	if runRec.Code != http.StatusOK {
		t.Fatalf("run-saved status = %d, want 200; body=%s", runRec.Code, runRec.Body.String())
	}
	var res report.Result
	json.Unmarshal(runRec.Body.Bytes(), &res)
	got := rowsByGroup(res)
	if got["new"].Count != 1 || got["working"].Count != 1 {
		t.Fatalf("rows = %+v, want new=1 working=1", res.Rows)
	}
}
