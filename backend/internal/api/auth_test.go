package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"saltcrm/internal/agent"
	"saltcrm/internal/identity"
)

func TestLoginSetsSessionCookieAndMeReturnsUser(t *testing.T) {
	db, router := newTestRouter(t)
	hash, _ := identity.HashPassword("s3cret")
	db.Create(&agent.Agent{Name: "Sam Carter", Email: "sam@isp.example", PasswordHash: hash, Role: agent.RoleAdmin})

	// Log in.
	loginReq := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"email":"sam@isp.example","password":"s3cret"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d; body=%s", loginRec.Code, http.StatusOK, loginRec.Body.String())
	}
	cookie := loginRec.Header().Get("Set-Cookie")
	if !strings.Contains(cookie, sessionCookie+"=") || !strings.Contains(cookie, "HttpOnly") {
		t.Fatalf("login Set-Cookie = %q, want an HttpOnly %s cookie", cookie, sessionCookie)
	}

	// The session cookie authenticates /auth/me.
	meReq := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	meReq.Header.Set("Cookie", cookie)
	meRec := httptest.NewRecorder()
	router.ServeHTTP(meRec, meReq)
	if meRec.Code != http.StatusOK {
		t.Fatalf("me status = %d, want %d; body=%s", meRec.Code, http.StatusOK, meRec.Body.String())
	}
	if !strings.Contains(meRec.Body.String(), "sam@isp.example") {
		t.Errorf("me body = %s, want the logged-in user", meRec.Body.String())
	}
	if strings.Contains(meRec.Body.String(), "PasswordHash") || strings.Contains(meRec.Body.String(), hash) {
		t.Errorf("me body leaked the password hash: %s", meRec.Body.String())
	}
}

func TestLoginWrongPasswordReturns401(t *testing.T) {
	db, router := newTestRouter(t)
	hash, _ := identity.HashPassword("s3cret")
	db.Create(&agent.Agent{Name: "Sam Carter", Email: "sam@isp.example", PasswordHash: hash, Role: agent.RoleAdmin})

	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"email":"sam@isp.example","password":"wrong"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestMeWithoutSessionReturns401(t *testing.T) {
	_, router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestLogoutInvalidatesSession(t *testing.T) {
	db, router := newTestRouter(t)
	hash, _ := identity.HashPassword("s3cret")
	db.Create(&agent.Agent{Name: "Sam Carter", Email: "sam@isp.example", PasswordHash: hash, Role: agent.RoleAdmin})

	loginReq := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"email":"sam@isp.example","password":"s3cret"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	cookie := loginRec.Header().Get("Set-Cookie")

	logoutReq := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	logoutReq.Header.Set("Cookie", cookie)
	logoutRec := httptest.NewRecorder()
	router.ServeHTTP(logoutRec, logoutReq)
	if logoutRec.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d, want %d", logoutRec.Code, http.StatusNoContent)
	}

	// The original cookie no longer authenticates.
	meReq := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	meReq.Header.Set("Cookie", cookie)
	meRec := httptest.NewRecorder()
	router.ServeHTTP(meRec, meReq)
	if meRec.Code != http.StatusUnauthorized {
		t.Errorf("me after logout status = %d, want %d", meRec.Code, http.StatusUnauthorized)
	}
}
