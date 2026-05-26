package api

import (
	"net/http"
	"strconv"

	"ispcrm/internal/subscription"

	"github.com/gin-gonic/gin"
)

type subscriptionHandler struct {
	svc *subscription.Service
}

func (h *subscriptionHandler) listForCustomer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	subs, err := h.svc.ListForCustomer(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list subscriptions"})
		return
	}
	if subs == nil {
		subs = []subscription.Subscription{}
	}
	c.JSON(http.StatusOK, subs)
}
