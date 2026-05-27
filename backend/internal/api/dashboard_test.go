package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"saltcrm/internal/activity"
	"saltcrm/internal/agent"
	"saltcrm/internal/lead"
	"saltcrm/internal/opportunity"
	"saltcrm/internal/supportcase"

	"gorm.io/gorm"
)

// meID looks up the id of the agent loginAs created for an email.
func meID(t *testing.T, db *gorm.DB, email string) uint {
	t.Helper()
	var a agent.Agent
	if err := db.Where("email = ?", email).First(&a).Error; err != nil {
		t.Fatalf("look up %s: %v", email, err)
	}
	return a.ID
}

func TestDashboardRequiresAuth(t *testing.T) {
	_, router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d (auth required)", rec.Code, http.StatusUnauthorized)
	}
}

func TestDashboardMyOpenCases(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "agent@isp.example", agent.RoleAgent, nil)
	me := meID(t, db, "agent@isp.example")
	other := agent.Agent{Name: "Other", Email: "other@isp.example"}
	db.Create(&other)

	mk := func(subject string, status supportcase.Status, assignee *uint) {
		db.Omit("AssignedAgent").Create(&supportcase.Case{
			CustomerID: 1, Subject: subject, Category: supportcase.CategoryGeneral,
			Priority: supportcase.PriorityLow, Status: status, AssignedAgentID: assignee,
		})
	}
	mk("mine, open", supportcase.StatusOpen, &me)
	mk("mine, closed", supportcase.StatusClosed, &me)
	mk("someone else's, open", supportcase.StatusOpen, &other.ID)

	rec := authGet(t, router, "/dashboard", cookie)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var dash struct {
		MyOpenCases []supportcase.Case `json:"myOpenCases"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &dash); err != nil {
		t.Fatalf("decode dashboard: %v; body=%s", err, rec.Body.String())
	}
	if len(dash.MyOpenCases) != 1 || dash.MyOpenCases[0].Subject != "mine, open" {
		t.Errorf("myOpenCases = %+v, want only the agent's active case", dash.MyOpenCases)
	}
}

func TestDashboardMyTasks(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "agent@isp.example", agent.RoleAgent, nil)
	me := meID(t, db, "agent@isp.example")
	other := agent.Agent{Name: "Other", Email: "other@isp.example"}
	db.Create(&other)

	now := time.Now()
	mk := func(typ activity.Type, st activity.Status, user uint, subj string) {
		u := user
		db.Create(&activity.Activity{
			Type: typ, Status: st, Subject: subj, AssignedUserID: &u,
			ParentType: "account", ParentID: 1, OccurredAt: now,
		})
	}
	mk(activity.TypeTask, activity.StatusOpen, me, "my open task")
	mk(activity.TypeTask, activity.StatusDone, me, "my done task")     // excluded
	mk(activity.TypeCall, activity.StatusOpen, me, "my call")          // excluded
	mk(activity.TypeTask, activity.StatusOpen, other.ID, "their task") // excluded

	rec := authGet(t, router, "/dashboard", cookie)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var dash struct {
		MyTasks []activity.Activity `json:"myTasks"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &dash); err != nil {
		t.Fatalf("decode dashboard: %v; body=%s", err, rec.Body.String())
	}
	if len(dash.MyTasks) != 1 || dash.MyTasks[0].Subject != "my open task" {
		t.Errorf("myTasks = %+v, want only the agent's open task", dash.MyTasks)
	}
}

func TestDashboardRecentLeadsVisibleNewestFirstCapped(t *testing.T) {
	db, router := newTestRouter(t)
	team := uint(10)
	cookie := loginAs(t, db, router, "agent@isp.example", agent.RoleAgent, &team)

	// Six leads visible via the user's team, created in ascending id order.
	for i := 1; i <= 6; i++ {
		db.Create(&lead.Lead{Name: fmt.Sprintf("Visible %d", i), Status: lead.StatusNew, TeamID: &team})
	}
	otherTeam := uint(20)
	db.Create(&lead.Lead{Name: "Invisible", Status: lead.StatusNew, TeamID: &otherTeam})

	rec := authGet(t, router, "/dashboard", cookie)
	var dash struct {
		RecentLeads []lead.Lead `json:"recentLeads"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &dash); err != nil {
		t.Fatalf("decode: %v; body=%s", err, rec.Body.String())
	}
	if len(dash.RecentLeads) != 5 {
		t.Fatalf("recentLeads = %d, want 5 (capped); %+v", len(dash.RecentLeads), dash.RecentLeads)
	}
	if dash.RecentLeads[0].Name != "Visible 6" {
		t.Errorf("first lead = %q, want newest 'Visible 6'", dash.RecentLeads[0].Name)
	}
	for _, l := range dash.RecentLeads {
		if l.Name == "Invisible" {
			t.Errorf("leaked a lead outside the user's visibility")
		}
	}
}

func TestDashboardPipelineByStageScoped(t *testing.T) {
	db, router := newTestRouter(t)
	team := uint(10)
	cookie := loginAs(t, db, router, "agent@isp.example", agent.RoleAgent, &team)

	db.Create(&opportunity.Opportunity{Name: "A", Stage: opportunity.StageProspecting, Amount: 100, TeamID: &team})
	db.Create(&opportunity.Opportunity{Name: "B", Stage: opportunity.StageProspecting, Amount: 200, TeamID: &team})
	otherTeam := uint(20)
	db.Create(&opportunity.Opportunity{Name: "C", Stage: opportunity.StageProspecting, Amount: 999, TeamID: &otherTeam})

	rec := authGet(t, router, "/dashboard", cookie)
	var dash struct {
		PipelineByStage []struct {
			Stage       string  `json:"stage"`
			Count       int     `json:"count"`
			TotalAmount float64 `json:"totalAmount"`
		} `json:"pipelineByStage"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &dash); err != nil {
		t.Fatalf("decode: %v; body=%s", err, rec.Body.String())
	}
	var prospecting *struct {
		Stage       string  `json:"stage"`
		Count       int     `json:"count"`
		TotalAmount float64 `json:"totalAmount"`
	}
	for i := range dash.PipelineByStage {
		if dash.PipelineByStage[i].Stage == "prospecting" {
			prospecting = &dash.PipelineByStage[i]
		}
	}
	if prospecting == nil {
		t.Fatalf("no prospecting stage in pipeline; got %+v", dash.PipelineByStage)
	}
	if prospecting.Count != 2 || prospecting.TotalAmount != 300 {
		t.Errorf("prospecting = {count:%d total:%.0f}, want {2 300} (invisible C excluded)", prospecting.Count, prospecting.TotalAmount)
	}
}
