package api

import (
	"errors"
	"net/http"
	"strconv"

	"saltcrm/internal/supportcase"

	"github.com/gin-gonic/gin"
)

type caseHandler struct {
	svc *supportcase.Service
}

func (h *caseHandler) listForCustomer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	cases, err := h.svc.ListForCustomer(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list cases"})
		return
	}
	// Always serialize as a JSON array, never null, even when empty.
	if cases == nil {
		cases = []supportcase.Case{}
	}
	c.JSON(http.StatusOK, cases)
}

// createCaseRequest is the body of POST /customers/:id/cases. Status is not
// accepted — a new case always starts Open.
type createCaseRequest struct {
	Subject     string               `json:"subject"`
	Description string               `json:"description"`
	Category    supportcase.Category `json:"category"`
	Priority    supportcase.Priority `json:"priority"`
}

func (h *caseHandler) create(c *gin.Context) {
	custID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	var req createCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	created, err := h.svc.Create(c.Request.Context(), supportcase.Case{
		CustomerID:  uint(custID),
		Subject:     req.Subject,
		Description: req.Description,
		Category:    req.Category,
		Priority:    req.Priority,
	})
	if err != nil {
		switch {
		case errors.Is(err, supportcase.ErrSubjectRequired),
			errors.Is(err, supportcase.ErrInvalidCategory),
			errors.Is(err, supportcase.ErrInvalidPriority):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create case"})
		}
		return
	}
	c.JSON(http.StatusCreated, created)
}

// addCommentRequest is the body of POST /cases/:id/comments.
type addCommentRequest struct {
	Body          string `json:"body"`
	AuthorAgentID uint   `json:"authorAgentId"`
}

func (h *caseHandler) addComment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}
	var req addCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	comment, err := h.svc.AddComment(c.Request.Context(), uint(id), req.AuthorAgentID, req.Body)
	if err != nil {
		switch {
		case errors.Is(err, supportcase.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "case not found"})
		case errors.Is(err, supportcase.ErrCommentBodyRequired):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add comment"})
		}
		return
	}
	c.JSON(http.StatusCreated, comment)
}

// patchCaseRequest is the body of PATCH /cases/:id. Every field is optional;
// only the provided ones are applied. Status changes go through the lifecycle
// state machine; priority/category/assignee are metadata edits.
type patchCaseRequest struct {
	Status          *supportcase.Status   `json:"status"`
	Priority        *supportcase.Priority `json:"priority"`
	Category        *supportcase.Category `json:"category"`
	AssignedAgentID *uint                 `json:"assignedAgentId"`
}

// respondCaseError maps a case service error to its HTTP status.
func respondCaseError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, supportcase.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "case not found"})
	case errors.Is(err, supportcase.ErrIllegalTransition):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, supportcase.ErrInvalidPriority), errors.Is(err, supportcase.ErrInvalidCategory):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update case"})
	}
}

func (h *caseHandler) patch(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}
	var req patchCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	ctx := c.Request.Context()
	caseID := uint(id)
	var updated supportcase.Case
	acted := false

	if req.Status != nil {
		updated, err = h.svc.ChangeStatus(ctx, caseID, *req.Status)
		if err != nil {
			respondCaseError(c, err)
			return
		}
		acted = true
	}
	if req.Priority != nil || req.Category != nil || req.AssignedAgentID != nil {
		updated, err = h.svc.UpdateMetadata(ctx, caseID, supportcase.MetadataPatch{
			Priority:        req.Priority,
			Category:        req.Category,
			AssignedAgentID: req.AssignedAgentID,
		})
		if err != nil {
			respondCaseError(c, err)
			return
		}
		acted = true
	}
	if !acted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *caseHandler) get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}
	kase, err := h.svc.Get(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, supportcase.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "case not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get case"})
		return
	}
	c.JSON(http.StatusOK, kase)
}
