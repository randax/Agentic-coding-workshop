package api

import (
	"errors"
	"net/http"
	"strconv"

	"saltcrm/internal/conversion"
	"saltcrm/internal/lead"

	"github.com/gin-gonic/gin"
)

type conversionHandler struct {
	svc *conversion.Service
}

func (h *conversionHandler) convert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lead id"})
		return
	}
	res, err := h.svc.Convert(c.Request.Context(), uint(id))
	if err != nil {
		switch {
		case errors.Is(err, lead.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "lead not found"})
		case errors.Is(err, conversion.ErrAlreadyConverted):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to convert lead"})
		}
		return
	}
	c.JSON(http.StatusOK, res)
}
