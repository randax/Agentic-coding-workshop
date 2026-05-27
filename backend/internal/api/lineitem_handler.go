package api

import (
	"errors"
	"net/http"
	"strconv"

	"saltcrm/internal/opportunity"
	"saltcrm/internal/product"

	"github.com/gin-gonic/gin"
)

type lineItemHandler struct {
	svc      *opportunity.LineItemService
	products *product.Service
}

func (h *lineItemHandler) list(c *gin.Context) {
	oppID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid opportunity id"})
		return
	}
	items, err := h.svc.ListByOpportunity(c.Request.Context(), uint(oppID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list line items"})
		return
	}
	if items == nil {
		items = []opportunity.LineItem{}
	}
	c.JSON(http.StatusOK, items)
}

func (h *lineItemHandler) add(c *gin.Context) {
	oppID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid opportunity id"})
		return
	}
	var body struct {
		ProductID uint `json:"productId"`
		Quantity  int  `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	// Snapshot the product's name and price at the time it's added to the deal.
	p, err := h.products.Get(c.Request.Context(), body.ProductID)
	if err != nil {
		if errors.Is(err, product.ErrNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unknown product"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load product"})
		return
	}
	created, err := h.svc.Add(c.Request.Context(), opportunity.LineItem{
		OpportunityID: uint(oppID),
		ProductID:     p.ID,
		ProductName:   p.Name,
		UnitPrice:     p.MonthlyPrice,
		Quantity:      body.Quantity,
	})
	if err != nil {
		if errors.Is(err, opportunity.ErrInvalidQuantity) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add line item"})
		return
	}
	c.JSON(http.StatusCreated, created)
}
