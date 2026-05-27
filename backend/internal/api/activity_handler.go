package api

import (
	"errors"
	"net/http"
	"strconv"

	"saltcrm/internal/access"
	"saltcrm/internal/activity"

	"github.com/gin-gonic/gin"
)

type activityHandler struct {
	svc *activity.Service
}

// list serves activities. With ?parentType=&parentId= it returns that record's
// timeline; otherwise all activities (the /m/activities module). Results are
// scoped to the current user's visibility.
func (h *activityHandler) list(c *gin.Context) {
	var (
		items []activity.Activity
		err   error
	)
	parentType := c.Query("parentType")
	if parentType != "" {
		parentID, perr := strconv.ParseUint(c.Query("parentId"), 10, 64)
		if perr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parentId"})
			return
		}
		items, err = h.svc.ListForParent(c.Request.Context(), parentType, uint(parentID))
	} else {
		items, err = h.svc.List(c.Request.Context())
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list activities"})
		return
	}
	if user, ok := currentUser(c); ok {
		visible := items[:0]
		for _, a := range items {
			if access.Visible(user, a.AssignedUserID, a.TeamID) {
				visible = append(visible, a)
			}
		}
		items = visible
	}
	if items == nil {
		items = []activity.Activity{}
	}
	c.JSON(http.StatusOK, items)
}

// log records a new activity.
func (h *activityHandler) log(c *gin.Context) {
	var a activity.Activity
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	// You own what you log: default the assignee/team to the current user so the
	// activity is visible to them (and their team) under the access rules.
	if user, ok := currentUser(c); ok {
		if a.AssignedUserID == nil {
			id := user.ID
			a.AssignedUserID = &id
		}
		if a.TeamID == nil {
			a.TeamID = user.TeamID
		}
	}
	created, err := h.svc.Log(c.Request.Context(), a)
	if err != nil {
		if errors.Is(err, activity.ErrSubjectRequired) || errors.Is(err, activity.ErrInvalidType) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log activity"})
		return
	}
	c.JSON(http.StatusCreated, created)
}

// complete marks a task activity done.
func (h *activityHandler) complete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid activity id"})
		return
	}
	done, err := h.svc.Complete(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, activity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "activity not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to complete activity"})
		return
	}
	c.JSON(http.StatusOK, done)
}
