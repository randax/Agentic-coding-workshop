package api

import (
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
