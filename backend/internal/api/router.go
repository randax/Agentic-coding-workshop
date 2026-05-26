// Package api wires the HTTP layer: it translates HTTP requests into service
// calls and serializes the results. It holds no business logic.
package api

import (
	"errors"
	"net/http"
	"strconv"

	"ispcrm/internal/customer"
	"ispcrm/internal/product"
	"ispcrm/internal/subscription"

	"github.com/gin-gonic/gin"
)

// NewRouter builds the HTTP handler for the CRM API.
func NewRouter(
	customers *customer.Service,
	products *product.Service,
	subscriptions *subscription.Service,
) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	ch := &customerHandler{svc: customers}
	r.GET("/customers", ch.list)
	r.GET("/customers/:id", ch.get)

	ph := &productHandler{svc: products}
	r.GET("/products", ph.list)
	r.POST("/products", ph.create)
	r.POST("/products/:id/retire", ph.retire)

	sh := &subscriptionHandler{svc: subscriptions}
	r.GET("/customers/:id/subscriptions", sh.listForCustomer)
	r.POST("/customers/:id/subscriptions", sh.assign)
	r.POST("/subscriptions/:id/cancel", sh.cancel)

	return r
}

// corsMiddleware allows the Next.js dev frontend to call the API from the browser.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

type customerHandler struct {
	svc *customer.Service
}

func (h *customerHandler) list(c *gin.Context) {
	customers, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list customers"})
		return
	}
	// Always serialize as a JSON array, never null, even when empty.
	if customers == nil {
		customers = []customer.Customer{}
	}
	c.JSON(http.StatusOK, customers)
}

func (h *customerHandler) get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	cust, err := h.svc.Get(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, customer.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get customer"})
		return
	}
	c.JSON(http.StatusOK, cust)
}
