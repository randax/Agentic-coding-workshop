// Package api wires the HTTP layer: it translates HTTP requests into service
// calls and serializes the results. It holds no business logic.
package api

import (
	"errors"
	"net/http"
	"os"
	"strconv"

	"saltcrm/internal/agent"
	"saltcrm/internal/customer"
	"saltcrm/internal/identity"
	"saltcrm/internal/product"
	"saltcrm/internal/subscription"
	"saltcrm/internal/supportcase"

	"github.com/gin-gonic/gin"
)

// NewRouter builds the HTTP handler for the CRM API.
func NewRouter(
	customers *customer.Service,
	products *product.Service,
	subscriptions *subscription.Service,
	agents *agent.Service,
	cases *supportcase.Service,
	identitySvc *identity.Service,
) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	auth := &authHandler{svc: identitySvc}
	r.POST("/auth/login", auth.login)
	r.POST("/auth/logout", auth.logout)
	r.GET("/auth/me", auth.me)

	ch := &customerHandler{svc: customers}
	r.GET("/customers", ch.list)
	r.POST("/customers", ch.create)
	r.GET("/customers/:id", ch.get)
	r.PUT("/customers/:id", ch.update)

	// "Accounts" is the SaltCRM-facing name for customers; the generic /m/accounts
	// views read records here. Legacy /customers routes above stay for now.
	r.GET("/accounts", ch.list)
	r.GET("/accounts/:id", ch.get)
	r.PUT("/accounts/:id", ch.update)

	mh := &metadataHandler{reg: defaultRegistry()}
	r.GET("/metadata/:module", mh.get)

	ph := &productHandler{svc: products}
	r.GET("/products", ph.list)
	r.GET("/products/:id", ph.get)
	r.POST("/products", ph.create)
	r.PUT("/products/:id", ph.update)
	r.POST("/products/:id/retire", ph.retire)
	r.POST("/products/:id/unretire", ph.unretire)

	sh := &subscriptionHandler{svc: subscriptions}
	r.GET("/customers/:id/subscriptions", sh.listForCustomer)
	r.POST("/customers/:id/subscriptions", sh.assign)
	r.POST("/subscriptions/:id/cancel", sh.cancel)

	ah := &agentHandler{svc: agents}
	r.GET("/agents", ah.list)

	caseH := &caseHandler{svc: cases}
	r.GET("/customers/:id/cases", caseH.listForCustomer)
	r.POST("/customers/:id/cases", caseH.create)
	r.GET("/cases/:id", caseH.get)
	r.PATCH("/cases/:id", caseH.patch)
	r.POST("/cases/:id/comments", caseH.addComment)

	return r
}

// corsMiddleware allows the Next.js dev frontend to call the API from the
// browser with credentials (the session cookie). Because credentials are
// allowed, the origin must be explicit (not "*"); it defaults to the Next dev
// server and can be overridden with SALTCRM_CORS_ORIGIN.
func corsMiddleware() gin.HandlerFunc {
	origin := os.Getenv("SALTCRM_CORS_ORIGIN")
	if origin == "" {
		origin = "http://localhost:3000"
	}
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
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
	filter := customer.Filter{
		Search: c.Query("search"),
		Status: customer.Status(c.Query("status")),
	}
	customers, err := h.svc.List(c.Request.Context(), filter)
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

// isValidationError reports whether err is one of the customer service's
// required-field or status validation errors (which map to HTTP 400).
func isValidationError(err error) bool {
	return errors.Is(err, customer.ErrNameRequired) ||
		errors.Is(err, customer.ErrEmailRequired) ||
		errors.Is(err, customer.ErrAccountNumberRequired) ||
		errors.Is(err, customer.ErrInvalidStatus)
}

func (h *customerHandler) create(c *gin.Context) {
	var cust customer.Customer
	if err := c.ShouldBindJSON(&cust); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	created, err := h.svc.Create(c.Request.Context(), cust)
	if err != nil {
		if isValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create customer"})
		return
	}
	c.JSON(http.StatusCreated, created)
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

func (h *customerHandler) update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	var cust customer.Customer
	if err := c.ShouldBindJSON(&cust); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	cust.ID = uint(id) // the URL is the source of truth for which customer to edit
	updated, err := h.svc.Update(c.Request.Context(), cust)
	if err != nil {
		if errors.Is(err, customer.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
			return
		}
		if isValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update customer"})
		return
	}
	c.JSON(http.StatusOK, updated)
}
