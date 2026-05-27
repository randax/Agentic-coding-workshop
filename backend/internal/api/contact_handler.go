package api

import (
	"errors"
	"net/http"
	"strconv"

	"saltcrm/internal/access"
	"saltcrm/internal/contact"

	"github.com/gin-gonic/gin"
)

type contactHandler struct {
	svc *contact.Service
}

func (h *contactHandler) list(c *gin.Context) {
	contacts, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list contacts"})
		return
	}
	if user, ok := currentUser(c); ok {
		visible := contacts[:0]
		for _, ct := range contacts {
			if access.Visible(user, ct.AssignedUserID, ct.TeamID) {
				visible = append(visible, ct)
			}
		}
		contacts = visible
	}
	if contacts == nil {
		contacts = []contact.Contact{}
	}
	c.JSON(http.StatusOK, contacts)
}

func (h *contactHandler) get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contact id"})
		return
	}
	ct, err := h.svc.Get(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, contact.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "contact not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get contact"})
		return
	}
	c.JSON(http.StatusOK, ct)
}

func (h *contactHandler) update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contact id"})
		return
	}
	var ct contact.Contact
	if err := c.ShouldBindJSON(&ct); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	ct.ID = uint(id)
	updated, err := h.svc.Update(c.Request.Context(), ct)
	if err != nil {
		if errors.Is(err, contact.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "contact not found"})
			return
		}
		if errors.Is(err, contact.ErrNameRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update contact"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

// listForAccount serves an account's contacts (the Account → Contacts subpanel).
func (h *contactHandler) listForAccount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}
	contacts, err := h.svc.ListByAccount(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list contacts"})
		return
	}
	if contacts == nil {
		contacts = []contact.Contact{}
	}
	c.JSON(http.StatusOK, contacts)
}
