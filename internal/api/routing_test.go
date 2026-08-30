package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouting(t *testing.T) {
	t.Parallel()
	handler := newTestHandler(t)

	for _, tt := range []struct {
		name   string
		method string
		path   string
		status int
	}{
		{name: "wrong method", method: http.MethodPost, path: "/healthz", status: http.StatusMethodNotAllowed},
		{name: "unknown route", method: http.MethodGet, path: "/unknown", status: http.StatusNotFound},
		{name: "trailing slash", method: http.MethodGet, path: "/healthz/", status: http.StatusNotFound},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			response := performRequest(handler, tt.method, tt.path, "", true)
			assert.Equalf(t, tt.status, response.Code, "status = %d, want %d", response.Code, tt.status)
		})
	}
}

func TestApplicationError(t *testing.T) {
	t.Parallel()
	response := performRequest(newTestHandler(t), http.MethodPost, "/api/v1/notes", "{", true)

	require.Equal(t, http.StatusBadRequest, response.Code)
	responseContentType := response.Header().Get("Content-Type")
	assert.Equal(t, "application/json", responseContentType)

	var body Error
	decodeErr := json.NewDecoder(response.Body).Decode(&body)
	require.NoError(t, decodeErr)
	assert.Equal(t, "Invalid request body", body.Message)
}
