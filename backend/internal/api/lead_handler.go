package api

import (
	"errors"
	"net/http"
	"strconv"

	"saltcrm/internal/access"
	"saltcrm/internal/lead"

	"github.com/gin-gonic/gin"
)

type leadHandler struct {
	svc *lead.Service
}

func (h *leadHandler) list(c *gin.Context) {
	leads, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list leads"})
		return
	}
	if user, ok := currentUser(c); ok {
		visible := leads[:0]
		for _, l := range leads {
			if access.Visible(user, l.AssignedUserID, l.TeamID) {
				visible = append(visible, l)
			}
		}
		leads = visible
	}
	if leads == nil {
		leads = []lead.Lead{}
	}
	c.JSON(http.StatusOK, leads)
}

func (h *leadHandler) get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lead id"})
		return
	}
	l, err := h.svc.Get(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, lead.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "lead not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get lead"})
		return
	}
	c.JSON(http.StatusOK, l)
}

func (h *leadHandler) create(c *gin.Context) {
	var l lead.Lead
	if err := c.ShouldBindJSON(&l); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	l.ID = 0
	defaultOwner(c, &l.AssignedUserID, &l.TeamID)
	created, err := h.svc.Create(c.Request.Context(), l)
	if err != nil {
		if errors.Is(err, lead.ErrNameRequired) || errors.Is(err, lead.ErrInvalidStatus) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create lead"})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *leadHandler) update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lead id"})
		return
	}
	var l lead.Lead
	if err := c.ShouldBindJSON(&l); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	l.ID = uint(id)
	updated, err := h.svc.Update(c.Request.Context(), l)
	if err != nil {
		if errors.Is(err, lead.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "lead not found"})
			return
		}
		if errors.Is(err, lead.ErrNameRequired) || errors.Is(err, lead.ErrInvalidStatus) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update lead"})
		return
	}
	c.JSON(http.StatusOK, updated)
}
