package api

import (
	"errors"
	"net/http"
	"strconv"

	"saltcrm/internal/access"
	"saltcrm/internal/opportunity"

	"github.com/gin-gonic/gin"
)

type opportunityHandler struct {
	svc *opportunity.Service
}

// visibleFor returns the opportunities the request's user may see.
func (h *opportunityHandler) visibleFor(c *gin.Context) ([]opportunity.Opportunity, error) {
	opps, err := h.svc.List(c.Request.Context())
	if err != nil {
		return nil, err
	}
	if user, ok := currentUser(c); ok {
		visible := opps[:0]
		for _, o := range opps {
			if access.Visible(user, o.AssignedUserID, o.TeamID) {
				visible = append(visible, o)
			}
		}
		opps = visible
	}
	return opps, nil
}

func (h *opportunityHandler) list(c *gin.Context) {
	opps, err := h.visibleFor(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list opportunities"})
		return
	}
	if opps == nil {
		opps = []opportunity.Opportunity{}
	}
	c.JSON(http.StatusOK, opps)
}

func (h *opportunityHandler) get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid opportunity id"})
		return
	}
	o, err := h.svc.Get(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, opportunity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "opportunity not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get opportunity"})
		return
	}
	c.JSON(http.StatusOK, o)
}

func (h *opportunityHandler) create(c *gin.Context) {
	var o opportunity.Opportunity
	if err := c.ShouldBindJSON(&o); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	o.ID = 0
	defaultOwner(c, &o.AssignedUserID, &o.TeamID)
	created, err := h.svc.Create(c.Request.Context(), o)
	if err != nil {
		if errors.Is(err, opportunity.ErrNameRequired) || errors.Is(err, opportunity.ErrInvalidAmount) || errors.Is(err, opportunity.ErrInvalidStage) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create opportunity"})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *opportunityHandler) update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid opportunity id"})
		return
	}
	var o opportunity.Opportunity
	if err := c.ShouldBindJSON(&o); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	o.ID = uint(id)
	updated, err := h.svc.Update(c.Request.Context(), o)
	if err != nil {
		if errors.Is(err, opportunity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "opportunity not found"})
			return
		}
		if errors.Is(err, opportunity.ErrNameRequired) || errors.Is(err, opportunity.ErrInvalidAmount) || errors.Is(err, opportunity.ErrInvalidStage) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update opportunity"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

// pipelineStage is one column of the pipeline view: a stage with its rolled-up
// count and total amount.
type pipelineStage struct {
	Stage       opportunity.Stage         `json:"stage"`
	Count       int                       `json:"count"`
	TotalAmount float64                   `json:"totalAmount"`
	Items       []opportunity.Opportunity `json:"items"`
}

// buildPipeline groups opportunities by stage in pipeline order, with a per-stage
// count and total. Shared by the pipeline endpoint and the dashboard dashlet.
func buildPipeline(opps []opportunity.Opportunity) []pipelineStage {
	byStage := map[opportunity.Stage]*pipelineStage{}
	stages := make([]*pipelineStage, 0, len(opportunity.Stages))
	for _, s := range opportunity.Stages {
		ps := &pipelineStage{Stage: s, Items: []opportunity.Opportunity{}}
		byStage[s] = ps
		stages = append(stages, ps)
	}
	for _, o := range opps {
		if ps, ok := byStage[o.Stage]; ok {
			ps.Count++
			ps.TotalAmount += o.Amount
			ps.Items = append(ps.Items, o)
		}
	}
	out := make([]pipelineStage, len(stages))
	for i, ps := range stages {
		out[i] = *ps
	}
	return out
}

// pipeline groups the visible opportunities by stage, in pipeline order.
func (h *opportunityHandler) pipeline(c *gin.Context) {
	opps, err := h.visibleFor(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load pipeline"})
		return
	}
	c.JSON(http.StatusOK, buildPipeline(opps))
}
