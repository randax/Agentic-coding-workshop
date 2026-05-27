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

func (h *studioHandler) listLayouts(c *gin.Context) {
	layouts, err := h.svc.GetLayouts(c.Request.Context(), c.Query("module"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list layouts"})
		return
	}
	if layouts == nil {
		layouts = map[string][]string{}
	}
	c.JSON(http.StatusOK, layouts)
}

// saveLayouts upserts each provided view's layout for a module. The body is
// {"module": "...", "views": {"list": [...], "detail": [...], "edit": [...]}}.
func (h *studioHandler) saveLayouts(c *gin.Context) {
	var body struct {
		Module string              `json:"module"`
		Views  map[string][]string `json:"views"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	for view, fields := range body.Views {
		if err := h.svc.SetLayout(c.Request.Context(), body.Module, view, fields); err != nil {
			if errors.Is(err, studio.ErrModuleRequired) || errors.Is(err, studio.ErrInvalidView) {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save layout"})
			return
		}
	}
	c.Status(http.StatusNoContent)
}
