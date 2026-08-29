package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestAppWithAuth(t *testing.T) *app {
	t.Helper()
	a := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	a.cfg.AdminPassword = "secret123"
	a.cfg.SessionSecret = "test-session-secret"
	return a
}

func TestHandleLogin_GET(t *testing.T) {
	t.Parallel()
	app := newTestAppWithAuth(t)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/login", nil)
	w := httptest.NewRecorder()
	app.handleLogin(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleLogin_POST_CorrectPassword(t *testing.T) {
	t.Parallel()
	app := newTestAppWithAuth(t)

	form := url.Values{"password": {"secret123"}}
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	app.handleLogin(w, req)

	location := w.Header().Get("Location")
	require.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/", location)
	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == sessionCookieName {
			found = true
			assert.True(t, c.Secure, "session cookie should be Secure by default")
		}
	}
	assert.True(t, found, "expected session cookie to be set")
}

func TestHandleLogin_POST_LocalDevelopmentCookie(t *testing.T) {
	t.Parallel()
	app := newTestAppWithAuth(t)
	app.cfg.InsecureSessionCookie = true

	form := url.Values{"password": {"secret123"}}
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	app.handleLogin(w, req)

	for _, c := range w.Result().Cookies() {
		if c.Name == sessionCookieName {
			assert.False(t, c.Secure, "local development session cookie should not be Secure")
		}
	}
}

func TestHandleLogin_POST_WrongPassword(t *testing.T) {
	t.Parallel()
	app := newTestAppWithAuth(t)

	form := url.Values{"password": {"wrong"}}
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	app.handleLogin(w, req)

	body := w.Body.String()
	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, body, "Invalid password")
}

func TestHandleLogin_UnsupportedMethod(t *testing.T) {
	t.Parallel()
	app := newTestAppWithAuth(t)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/login", nil)
	w := httptest.NewRecorder()
	app.handleLogin(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleLogout(t *testing.T) {
	t.Parallel()
	app := newTestAppWithAuth(t)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/logout", nil)
	w := httptest.NewRecorder()
	app.handleLogout(w, req)

	location := w.Header().Get("Location")
	require.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/login", location)
	cookies := w.Result().Cookies()
	for _, c := range cookies {
		if c.Name == sessionCookieName {
			assert.Equal(t, -1, c.MaxAge)
		}
	}
}

func TestSessionMiddleware_NoCookie(t *testing.T) {
	t.Parallel()
	app := newTestAppWithAuth(t)

	handler := app.sessionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusSeeOther, w.Code)
}

func TestSessionMiddleware_InvalidCookie(t *testing.T) {
	t.Parallel()
	app := newTestAppWithAuth(t)

	handler := app.sessionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "invalid-token", Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusSeeOther, w.Code)
}

func TestSessionMiddleware_ValidCookie(t *testing.T) {
	t.Parallel()
	app := newTestAppWithAuth(t)

	handler := app.sessionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	token := signSession(app.cfg.SessionSecret)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token, Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
