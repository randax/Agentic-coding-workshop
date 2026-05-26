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
