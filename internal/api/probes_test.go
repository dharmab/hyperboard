package api

import (
	"errors"
	"net/http"
	"testing"

	"github.com/dharmab/hyperboard/internal/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealth(t *testing.T) {
	t.Parallel()
	response := performRequest(newTestHandler(t), http.MethodGet, "/healthz", "", false)

	require.Equal(t, http.StatusOK, response.Code)
	responseContentType := response.Header().Get("Content-Type")
	assert.Equal(t, "application/json", responseContentType)
	responseBody := response.Body.String()
	assert.Equal(t, `"OK"`, responseBody)
}

func TestReadinessFailures(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name      string
		sqlErr    error
		mediaErr  error
		wantError string
	}{
		{name: "database", sqlErr: errors.New("database unavailable"), wantError: "Service is not ready"},
		{name: "object store", mediaErr: errors.New("storage unavailable"), wantError: "Service is not ready"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mediaStore := &faultInjectingMediaStore{MediaStore: memory.New(), pingErr: tt.mediaErr}
			server := NewServer(faultInjectingSQLStore{SQLStore: testSQLStore, pingErr: tt.sqlErr}, mediaStore)
			response := performRequest(newTestHandlerForServer(t, server), http.MethodGet, "/readyz", "", false)

			responseBody := response.Body.String()
			require.Equal(t, http.StatusServiceUnavailable, response.Code, "readiness body: %s", responseBody)
			assertJSONContentType(t, response.Header().Get("Content-Type"))
			assert.Contains(t, responseBody, tt.wantError, "readiness body")
			assert.NotContains(t, responseBody, "unavailable", "readiness response must not expose dependency errors")
		})
	}
}

func TestMetrics(t *testing.T) {
	t.Parallel()
	response := performRequest(newTestHandler(t), http.MethodGet, "/metrics", "", false)

	require.Equal(t, http.StatusOK, response.Code)
	responseContentType := response.Header().Get("Content-Type")
	assert.Empty(t, responseContentType)
	responseBody := response.Body.String()
	assert.Empty(t, responseBody)
}

func TestReadiness(t *testing.T) {
	t.Parallel()
	response := performRequest(newTestHandler(t), http.MethodGet, "/readyz", "", false)

	require.Equalf(t, http.StatusOK, response.Code, "status = %d, want %d", response.Code, http.StatusOK)
	contentType := response.Header().Get("Content-Type")
	assert.Equal(t, "application/json", contentType)
	responseBody := response.Body.String()
	assert.Equal(t, `"OK"`, responseBody)
}
