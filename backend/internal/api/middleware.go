package api

import (
	"net/http"

	"saltcrm/internal/agent"
	"saltcrm/internal/identity"

	"github.com/gin-gonic/gin"
)

const currentUserKey = "currentUser"

// requireAuth resolves the session cookie to the current user and stores it on
// the context, aborting with 401 if there is no valid session.
func requireAuth(identitySvc *identity.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(sessionCookie)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}
		user, err := identitySvc.CurrentUser(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}
		c.Set(currentUserKey, user)
		c.Next()
	}
}

// requireRole aborts with 403 unless the current user has one of the given roles.
// It must run after requireAuth.
func requireRole(roles ...agent.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}
		for _, r := range roles {
			if user.Role == r {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient role"})
	}
}

// currentUser returns the authenticated user stored by requireAuth, if any.
func currentUser(c *gin.Context) (agent.Agent, bool) {
	v, ok := c.Get(currentUserKey)
	if !ok {
		return agent.Agent{}, false
	}
	u, ok := v.(agent.Agent)
	return u, ok
}
