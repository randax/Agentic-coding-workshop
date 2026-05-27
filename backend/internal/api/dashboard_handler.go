package api

import (
	"net/http"
	"sort"

	"saltcrm/internal/access"
	"saltcrm/internal/activity"
	"saltcrm/internal/lead"
	"saltcrm/internal/opportunity"
	"saltcrm/internal/supportcase"

	"github.com/gin-gonic/gin"
)

// recentLeadsLimit caps how many recent leads the dashlet shows.
const recentLeadsLimit = 5

// dashboardHandler assembles the signed-in user's dashboard: a set of dashlets,
// each scoped to what that user may see. It composes existing domain services;
// it holds no business logic of its own.
type dashboardHandler struct {
	cases         *supportcase.Service
	activities    *activity.Service
	leads         *lead.Service
	opportunities *opportunity.Service
}

func (h *dashboardHandler) get(c *gin.Context) {
	user, ok := currentUser(c) // route is behind requireAuth, so this is set
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	ctx := c.Request.Context()
	fail := func() { c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load dashboard"}) }

	// My Open Cases: the user's active cases (assigned to them).
	myOpenCases, err := h.cases.ListOpenForAgent(ctx, user.ID)
	if err != nil {
		fail()
		return
	}

	// My Tasks: the user's open tasks.
	myTasks, err := h.activities.OpenTasksForUser(ctx, user.ID)
	if err != nil {
		fail()
		return
	}

	// Recent Leads: the visible leads, newest first, capped.
	allLeads, err := h.leads.List(ctx)
	if err != nil {
		fail()
		return
	}
	var recentLeads []lead.Lead
	for _, l := range allLeads {
		if access.Visible(user, l.AssignedUserID, l.TeamID) {
			recentLeads = append(recentLeads, l)
		}
	}
	sort.Slice(recentLeads, func(i, j int) bool { return recentLeads[i].ID > recentLeads[j].ID })
	if len(recentLeads) > recentLeadsLimit {
		recentLeads = recentLeads[:recentLeadsLimit]
	}

	// Pipeline by Stage: the visible opportunities, grouped by stage.
	allOpps, err := h.opportunities.List(ctx)
	if err != nil {
		fail()
		return
	}
	var visibleOpps []opportunity.Opportunity
	for _, o := range allOpps {
		if access.Visible(user, o.AssignedUserID, o.TeamID) {
			visibleOpps = append(visibleOpps, o)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"myOpenCases":     orEmpty(myOpenCases),
		"myTasks":         orEmpty(myTasks),
		"recentLeads":     orEmpty(recentLeads),
		"pipelineByStage": buildPipeline(visibleOpps),
	})
}

// orEmpty renders a nil slice as [] in JSON rather than null.
func orEmpty[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
