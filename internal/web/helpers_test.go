package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestApp(t *testing.T, handler http.Handler) *app {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	tmpls, err := parseTemplates()
	require.NoError(t, err, "failed to parse templates")

	api, err := newAPIClient(srv.URL, "test")
	require.NoError(t, err, "failed to create API client")

	return &app{
		cfg:   &config{},
		api:   api,
		media: newMediaClient(srv.URL, "test"),
		tmpls: tmpls,
	}
}

const postsAPIPath = "/api/v1/posts"

func jsonResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	data, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(data)
}
