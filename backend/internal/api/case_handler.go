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
