package web

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/posts", nil)
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	body, err := io.ReadAll(rec.Body)
	require.NoError(t, err)
	assert.JSONEq(t, `{"ok":true}`, string(body))
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
	require.NoError(t, err)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/tags/foo", nil)
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)

	assert.Equal(t, "/api/v1/tags/foo", gotPath)
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
	require.NoError(t, err)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/posts", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)

	assert.Equal(t, "Bearer test-token", gotAuth)
}

func TestAPIProxy_InvalidURL(t *testing.T) {
	t.Parallel()

	_, err := newAPIProxy("://bad")
	require.Error(t, err)
}
