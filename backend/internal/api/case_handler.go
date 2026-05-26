package api

import (
	"errors"
	"net/http"
	"strconv"

	"ispcrm/internal/supportcase"

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

// patchCaseRequest is the body of PATCH /cases/:id. For this slice only the
// status is editable (priority/category/assignee arrive in a later slice).
type patchCaseRequest struct {
	Status supportcase.Status `json:"status"`
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
	if req.Status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
		return
	}
	updated, err := h.svc.ChangeStatus(c.Request.Context(), uint(id), req.Status)
	if err != nil {
		switch {
		case errors.Is(err, supportcase.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "case not found"})
		case errors.Is(err, supportcase.ErrIllegalTransition):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update case"})
		}
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
