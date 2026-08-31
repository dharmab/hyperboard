package web

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type deadlineRecorder struct {
	*httptest.ResponseRecorder
	deadlines []time.Time
}

func (r *deadlineRecorder) SetReadDeadline(deadline time.Time) error {
	r.deadlines = append(r.deadlines, deadline)
	return nil
}

func TestLoginBodyReadDeadline(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		method       string
		target       string
		wantDeadline bool
	}{
		{name: "login submission", method: http.MethodPost, target: "/login", wantDeadline: true},
		{name: "login page", method: http.MethodGet, target: "/login", wantDeadline: false},
		{name: "authenticated upload", method: http.MethodPost, target: "/upload", wantDeadline: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recorder := &deadlineRecorder{ResponseRecorder: httptest.NewRecorder()}
			handler := loginBodyReadDeadline(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.wantDeadline && assert.Len(t, recorder.deadlines, 1) {
					assert.False(t, recorder.deadlines[0].IsZero())
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			req := httptest.NewRequestWithContext(t.Context(), tt.method, tt.target, nil)

			handler.ServeHTTP(recorder, req)

			assert.Equal(t, http.StatusNoContent, recorder.Code)
			if tt.wantDeadline {
				require.Len(t, recorder.deadlines, 2)
				assert.WithinDuration(t, time.Now().Add(loginBodyReadTimeout), recorder.deadlines[0], time.Second)
				assert.True(t, recorder.deadlines[1].IsZero(), "read deadline should be cleared after login")
			} else {
				assert.Empty(t, recorder.deadlines)
			}
		})
	}
}

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
