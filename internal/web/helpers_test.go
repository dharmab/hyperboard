package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestApp(t *testing.T, handler http.Handler) *app {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	tmpls, err := parseTemplates()
	if err != nil {
		t.Fatalf("failed to parse templates: %v", err)
	}

	api, err := newAPIClient(srv.URL, "test")
	if err != nil {
		t.Fatalf("failed to create API client: %v", err)
	}

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
