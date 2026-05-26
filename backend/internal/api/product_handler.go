package api

import (
	"errors"
	"net/http"
	"strconv"

	"ispcrm/internal/product"

	"github.com/gin-gonic/gin"
)

type productHandler struct {
	svc *product.Service
}

func (h *productHandler) list(c *gin.Context) {
	products, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list products"})
		return
	}
	if products == nil {
		products = []product.Product{}
	}
	c.JSON(http.StatusOK, products)
}

func (h *productHandler) create(c *gin.Context) {
	var p product.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	p.ID = 0 // the server assigns identifiers
	created, err := h.svc.Create(c.Request.Context(), p)
	if err != nil {
		if errors.Is(err, product.ErrInvalidCategory) || errors.Is(err, product.ErrInvalidPrice) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create product"})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *productHandler) update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	var p product.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	p.ID = uint(id) // the URL is the source of truth for which product to edit
	updated, err := h.svc.Update(c.Request.Context(), p)
	if err != nil {
		if errors.Is(err, product.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		if errors.Is(err, product.ErrInvalidCategory) || errors.Is(err, product.ErrInvalidPrice) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update product"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *productHandler) retire(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	if err := h.svc.Retire(c.Request.Context(), uint(id)); err != nil {
		if errors.Is(err, product.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retire product"})
		return
	}
	c.Status(http.StatusOK)
}

func (h *productHandler) unretire(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	if err := h.svc.Unretire(c.Request.Context(), uint(id)); err != nil {
		if errors.Is(err, product.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unretire product"})
		return
	}
	c.Status(http.StatusOK)
}
