package api

import (
	"net/http"

	"saltcrm/internal/agent"

	"github.com/gin-gonic/gin"
)

type agentHandler struct {
	svc *agent.Service
}

func (h *agentHandler) list(c *gin.Context) {
	agents, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list agents"})
		return
	}
	// Always serialize as a JSON array, never null, even when empty.
	if agents == nil {
		agents = []agent.Agent{}
	}
	c.JSON(http.StatusOK, agents)
}
