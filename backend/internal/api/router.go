// Package api wires the HTTP layer: it translates HTTP requests into service
// calls and serializes the results. It holds no business logic.
package api

import (
	"net/http"

	"ispcrm/internal/customer"

	"github.com/gin-gonic/gin"
)

// NewRouter builds the HTTP handler for the CRM API.
func NewRouter(customers *customer.Service) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	h := &customerHandler{svc: customers}
	r.GET("/customers", h.list)

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
