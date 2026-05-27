package api

import (
	"errors"
	"net/http"
	"strconv"

	"saltcrm/internal/conversion"

	"github.com/gin-gonic/gin"
)

// conversionHandler exposes the lead→account conversion endpoint. It is a thin
// translator: parse the lead id, resolve the current user, call the service, and
// map its typed errors to status codes.
type conversionHandler struct {
	svc *conversion.Service
}

// convert handles POST /leads/:id/convert. The request body is empty/{} for this
// slice (always a new account, no opportunity); the response is the new ids.
func (h *conversionHandler) convert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lead id"})
		return
	}
	// The route is behind requireAuth + requireRole, so a user is always set;
	// this guards against misconfiguration rather than a reachable request.
	user, ok := currentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	result, err := h.svc.Convert(c.Request.Context(), user, uint(id))
	if err != nil {
		switch {
		case errors.Is(err, conversion.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "lead not found"})
		case errors.Is(err, conversion.ErrNotQualified), errors.Is(err, conversion.ErrAlreadyConverted):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, conversion.ErrCompanyRequired):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to convert lead"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"accountId": result.AccountID, "contactId": result.ContactID})
}
