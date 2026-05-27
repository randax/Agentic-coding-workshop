package api

import (
	"errors"
	"net/http"

	"saltcrm/internal/identity"

	"github.com/gin-gonic/gin"
)

// sessionCookie is the name of the HTTP-only cookie holding the session token.
const sessionCookie = "saltcrm_session"

type authHandler struct {
	svc *identity.Service
}

func (h *authHandler) login(c *gin.Context) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	user, token, err := h.svc.Login(c.Request.Context(), body.Email, body.Password)
	if err != nil {
		if errors.Is(err, identity.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log in"})
		return
	}
	setSessionCookie(c, token, int((7 * 24 * 60 * 60)))
	c.JSON(http.StatusOK, user)
}

func (h *authHandler) logout(c *gin.Context) {
	if token, err := c.Cookie(sessionCookie); err == nil {
		_ = h.svc.Logout(c.Request.Context(), token)
	}
	setSessionCookie(c, "", -1)
	c.Status(http.StatusNoContent)
}

func (h *authHandler) me(c *gin.Context) {
	token, err := c.Cookie(sessionCookie)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	user, err := h.svc.CurrentUser(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// setSessionCookie writes the session cookie as HTTP-only and SameSite=Lax so it
// rides along with same-site frontend requests. Secure is omitted for local dev.
func setSessionCookie(c *gin.Context, token string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookie, token, maxAge, "/", "", false, true)
}
