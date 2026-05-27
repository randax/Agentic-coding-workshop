package api

import (
	"errors"
	"net/http"

	"saltcrm/internal/studio"

	"github.com/gin-gonic/gin"
)

type studioHandler struct {
	svc *studio.Service
}

func (h *studioHandler) listFields(c *gin.Context) {
	module := c.Query("module")
	defs, err := h.svc.ListByModule(c.Request.Context(), module)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list fields"})
		return
	}
	if defs == nil {
		defs = []studio.FieldDef{}
	}
	c.JSON(http.StatusOK, defs)
}

func (h *studioHandler) addField(c *gin.Context) {
	var d studio.FieldDef
	if err := c.ShouldBindJSON(&d); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	created, err := h.svc.AddField(c.Request.Context(), d)
	if err != nil {
		if errors.Is(err, studio.ErrModuleRequired) || errors.Is(err, studio.ErrNameRequired) ||
			errors.Is(err, studio.ErrLabelRequired) || errors.Is(err, studio.ErrInvalidType) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add field"})
		return
	}
	c.JSON(http.StatusCreated, created)
}
