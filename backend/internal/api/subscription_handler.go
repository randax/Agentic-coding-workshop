package api

import (
	"errors"
	"net/http"
	"strconv"

	"saltcrm/internal/product"
	"saltcrm/internal/subscription"

	"github.com/gin-gonic/gin"
)

type subscriptionHandler struct {
	svc *subscription.Service
}

// assignRequest is the body of POST /customers/:id/subscriptions.
type assignRequest struct {
	ProductID uint `json:"productId"`
	Quantity  int  `json:"quantity"`
}

// list serves all subscriptions (the generic /m/subscriptions module).
func (h *subscriptionHandler) list(c *gin.Context) {
	subs, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list subscriptions"})
		return
	}
	if subs == nil {
		subs = []subscription.Subscription{}
	}
	c.JSON(http.StatusOK, subs)
}

func (h *subscriptionHandler) assign(c *gin.Context) {
	custID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	var req assignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	sub, err := h.svc.Assign(c.Request.Context(), uint(custID), req.ProductID, req.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, subscription.ErrInvalidQuantity):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, product.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		case errors.Is(err, subscription.ErrProductRetired):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign subscription"})
		}
		return
	}
	c.JSON(http.StatusCreated, sub)
}

func (h *subscriptionHandler) cancel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription id"})
		return
	}
	sub, err := h.svc.Cancel(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, subscription.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel subscription"})
		return
	}
	c.JSON(http.StatusOK, sub)
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
