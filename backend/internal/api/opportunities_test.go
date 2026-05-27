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
)

func TestGetOpportunitiesRequiresAuth(t *testing.T) {
	_, router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/opportunities", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestPipelineGroupsByStageWithTotals(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "admin@isp.example", agent.RoleAdmin, nil)
	db.Create(&opportunity.Opportunity{Name: "A", AccountID: 1, Amount: 1000, Stage: opportunity.StageProspecting, Probability: 10})
	db.Create(&opportunity.Opportunity{Name: "B", AccountID: 1, Amount: 500, Stage: opportunity.StageProspecting, Probability: 10})
	db.Create(&opportunity.Opportunity{Name: "C", AccountID: 1, Amount: 2000, Stage: opportunity.StageProposal, Probability: 50})

	rec := authGet(t, router, "/opportunities/pipeline", cookie)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var stages []struct {
		Stage       string  `json:"stage"`
		Count       int     `json:"count"`
		TotalAmount float64 `json:"totalAmount"`
	}
	json.Unmarshal(rec.Body.Bytes(), &stages)
	if len(stages) != 6 {
		t.Fatalf("got %d stage columns, want 6 (the full pipeline)", len(stages))
	}
	byStage := map[string]struct {
		count int
		total float64
	}{}
	for _, s := range stages {
		byStage[s.Stage] = struct {
			count int
			total float64
		}{s.Count, s.TotalAmount}
	}
	if byStage["prospecting"].count != 2 || byStage["prospecting"].total != 1500 {
		t.Errorf("prospecting = %+v, want count 2 total 1500", byStage["prospecting"])
	}
	if byStage["proposal"].count != 1 || byStage["proposal"].total != 2000 {
		t.Errorf("proposal = %+v, want count 1 total 2000", byStage["proposal"])
	}
}

func TestPutOpportunityAsManagerSyncsProbability(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "manager@isp.example", agent.RoleManager, nil)
	o := opportunity.Opportunity{Name: "Deal", AccountID: 1, Amount: 1000, Stage: opportunity.StageProspecting, Probability: 10}
	db.Create(&o)

	req := httptest.NewRequest(http.MethodPut, "/opportunities/"+strconv.FormatUint(uint64(o.ID), 10),
		strings.NewReader(`{"name":"Deal","accountId":1,"amount":1000,"stage":"closed_won"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got map[string]any
	json.Unmarshal(rec.Body.Bytes(), &got)
	if got["stage"] != "closed_won" || got["probability"].(float64) != 100 {
		t.Errorf("got stage=%v prob=%v, want closed_won / 100", got["stage"], got["probability"])
	}
}

func TestPutOpportunityAsAgentIsForbidden(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "agent@isp.example", agent.RoleAgent, nil)
	o := opportunity.Opportunity{Name: "Deal", AccountID: 1, Amount: 1000, Stage: opportunity.StageProspecting}
	db.Create(&o)

	req := httptest.NewRequest(http.MethodPut, "/opportunities/"+strconv.FormatUint(uint64(o.ID), 10), strings.NewReader(`{"name":"X","accountId":1,"amount":1,"stage":"proposal"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}
