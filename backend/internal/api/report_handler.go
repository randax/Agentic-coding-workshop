package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"saltcrm/internal/access"
	"saltcrm/internal/agent"
	"saltcrm/internal/contact"
	"saltcrm/internal/customer"
	"saltcrm/internal/lead"
	"saltcrm/internal/opportunity"
	"saltcrm/internal/report"

	"github.com/gin-gonic/gin"
)

// errUnsupportedModule is returned when a report names a module the handler
// cannot supply records for.
var errUnsupportedModule = errors.New("unsupported report module")

// reportHandler runs and persists reports. Like the dashboard and search
// handlers, it composes existing domain services to gather the records a report
// runs over — restricted to what the signed-in user may see — and delegates the
// filtering/grouping/aggregation to the pure report package. It holds no
// reporting logic of its own.
type reportHandler struct {
	reports       *report.Service
	customers     *customer.Service
	contacts      *contact.Service
	leads         *lead.Service
	opportunities *opportunity.Service
}

// run executes an ad-hoc report definition from the request body.
func (h *reportHandler) run(c *gin.Context) {
	var def report.Definition
	if err := c.ShouldBindJSON(&def); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	h.runDefinition(c, def)
}

// save validates and persists a named report.
func (h *reportHandler) save(c *gin.Context) {
	var r report.Saved
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	saved, err := h.reports.Save(c.Request.Context(), r)
	if err != nil {
		if errors.Is(err, report.ErrNameRequired) || errors.Is(err, report.ErrModuleRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save report"})
		return
	}
	c.JSON(http.StatusCreated, saved)
}

// list returns the saved reports.
func (h *reportHandler) list(c *gin.Context) {
	saved, err := h.reports.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list reports"})
		return
	}
	if saved == nil {
		saved = []report.Saved{}
	}
	c.JSON(http.StatusOK, saved)
}

// get returns a single saved report.
func (h *reportHandler) get(c *gin.Context) {
	saved, ok := h.lookup(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, saved)
}

// runSaved re-runs a stored report by ID.
func (h *reportHandler) runSaved(c *gin.Context) {
	saved, ok := h.lookup(c)
	if !ok {
		return
	}
	h.runDefinition(c, saved.Definition)
}

// lookup resolves the :id path param to a saved report, writing the appropriate
// error response (400/404/500) and returning ok=false on failure.
func (h *reportHandler) lookup(c *gin.Context) (report.Saved, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report id"})
		return report.Saved{}, false
	}
	saved, err := h.reports.Get(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, report.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
			return report.Saved{}, false
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get report"})
		return report.Saved{}, false
	}
	return saved, true
}

// runDefinition gathers the visible records for the definition's module and runs
// the report over them.
func (h *reportHandler) runDefinition(c *gin.Context, def report.Definition) {
	user, ok := currentUser(c) // route is behind requireAuth, so this is set
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	records, err := h.recordsFor(c.Request.Context(), user, def.Module)
	if err != nil {
		if errors.Is(err, errUnsupportedModule) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported report module"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to run report"})
		return
	}
	c.JSON(http.StatusOK, report.Run(def, records))
}

// recordsFor loads a module's records as generic field bags, restricted to what
// the user may see (own-or-team; admins see all). Marshalling each typed record
// to JSON and back flattens custom fields onto the top level, so reports filter
// and group on custom fields exactly like core ones.
func (h *reportHandler) recordsFor(ctx context.Context, user agent.Agent, module string) ([]report.Record, error) {
	switch module {
	case "accounts":
		all, err := h.customers.List(ctx, customer.Filter{})
		if err != nil {
			return nil, err
		}
		var visible []customer.Customer
		for _, r := range all {
			if access.Visible(user, r.AssignedUserID, r.TeamID) {
				visible = append(visible, r)
			}
		}
		return toReportRecords(visible)
	case "contacts":
		all, err := h.contacts.List(ctx)
		if err != nil {
			return nil, err
		}
		var visible []contact.Contact
		for _, r := range all {
			if access.Visible(user, r.AssignedUserID, r.TeamID) {
				visible = append(visible, r)
			}
		}
		return toReportRecords(visible)
	case "leads":
		all, err := h.leads.List(ctx)
		if err != nil {
			return nil, err
		}
		var visible []lead.Lead
		for _, r := range all {
			if access.Visible(user, r.AssignedUserID, r.TeamID) {
				visible = append(visible, r)
			}
		}
		return toReportRecords(visible)
	case "opportunities":
		all, err := h.opportunities.List(ctx)
		if err != nil {
			return nil, err
		}
		var visible []opportunity.Opportunity
		for _, r := range all {
			if access.Visible(user, r.AssignedUserID, r.TeamID) {
				visible = append(visible, r)
			}
		}
		return toReportRecords(visible)
	default:
		return nil, errUnsupportedModule
	}
}

// toReportRecords converts a slice of typed records into generic field bags via
// JSON, so custom fields (flattened on marshal) become first-class report fields.
func toReportRecords(v any) ([]report.Record, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var records []report.Record
	if err := json.Unmarshal(b, &records); err != nil {
		return nil, err
	}
	return records, nil
}
