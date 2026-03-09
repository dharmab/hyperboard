package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func newTestAppWithAuth(mock *mockAPIClient) *App {
	app := newTestApp(mock)
	app.cfg.AdminPassword = "secret123"
	app.cfg.SessionSecret = "test-session-secret"
	return app
}

func TestHandleLogin_GET(t *testing.T) {
	t.Parallel()
	app := newTestAppWithAuth(&mockAPIClient{})

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	w := httptest.NewRecorder()
	app.handleLogin(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleLogin_POST_CorrectPassword(t *testing.T) {
	t.Parallel()
	app := newTestAppWithAuth(&mockAPIClient{})

	form := url.Values{"password": {"secret123"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	app.handleLogin(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusSeeOther)
	}
	if loc := w.Header().Get("Location"); loc != "/" {
		t.Errorf("Location = %q, want /", loc)
	}
	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == sessionCookieName {
			found = true
		}
	}
	if !found {
		t.Error("expected session cookie to be set")
	}
}

func TestHandleLogin_POST_WrongPassword(t *testing.T) {
	t.Parallel()
	app := newTestAppWithAuth(&mockAPIClient{})

	form := url.Values{"password": {"wrong"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	app.handleLogin(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), "Invalid password") {
		t.Error("expected error message in response body")
	}
}

func TestHandleLogin_UnsupportedMethod(t *testing.T) {
	t.Parallel()
	app := newTestAppWithAuth(&mockAPIClient{})

	req := httptest.NewRequest(http.MethodPut, "/login", nil)
	w := httptest.NewRecorder()
	app.handleLogin(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleLogout(t *testing.T) {
	t.Parallel()
	app := newTestAppWithAuth(&mockAPIClient{})

	req := httptest.NewRequest(http.MethodGet, "/logout", nil)
	w := httptest.NewRecorder()
	app.handleLogout(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusSeeOther)
	}
	if loc := w.Header().Get("Location"); loc != "/login" {
		t.Errorf("Location = %q, want /login", loc)
	}
	cookies := w.Result().Cookies()
	for _, c := range cookies {
		if c.Name == sessionCookieName && c.MaxAge != -1 {
			t.Errorf("session cookie MaxAge = %d, want -1", c.MaxAge)
		}
	}
}

func TestSessionMiddleware_NoCookie(t *testing.T) {
	t.Parallel()
	app := newTestAppWithAuth(&mockAPIClient{})

	handler := app.sessionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusSeeOther)
	}
}

func TestSessionMiddleware_InvalidCookie(t *testing.T) {
	t.Parallel()
	app := newTestAppWithAuth(&mockAPIClient{})

	handler := app.sessionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "invalid-token"})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusSeeOther)
	}
}

func TestSessionMiddleware_ValidCookie(t *testing.T) {
	t.Parallel()
	app := newTestAppWithAuth(&mockAPIClient{})

	handler := app.sessionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	token := signSession(app.cfg.SessionSecret)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}
