package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIProxy_ForwardsRequests(t *testing.T) {
	t.Parallel()

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(backend.Close)

	proxy, err := newAPIProxy(backend.URL)
	if err != nil {
		t.Fatalf("newAPIProxy: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/posts", nil)
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) != `{"ok":true}` {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestAPIProxy_PreservesPath(t *testing.T) {
	t.Parallel()

	var gotPath string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(backend.Close)

	proxy, err := newAPIProxy(backend.URL)
	if err != nil {
		t.Fatalf("newAPIProxy: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tags/foo", nil)
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)

	if gotPath != "/api/v1/tags/foo" {
		t.Fatalf("expected path /api/v1/tags/foo, got %s", gotPath)
	}
}

func TestAPIProxy_PreservesHeaders(t *testing.T) {
	t.Parallel()

	var gotAuth string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(backend.Close)

	proxy, err := newAPIProxy(backend.URL)
	if err != nil {
		t.Fatalf("newAPIProxy: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/posts", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)

	if gotAuth != "Bearer test-token" {
		t.Fatalf("expected Authorization header 'Bearer test-token', got %q", gotAuth)
	}
}

func TestAPIProxy_InvalidURL(t *testing.T) {
	t.Parallel()

	_, err := newAPIProxy("://bad")
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}
}
