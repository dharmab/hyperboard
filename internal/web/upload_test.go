package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleUpload_GET(t *testing.T) {
	t.Parallel()
	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/upload", nil)
	w := httptest.NewRecorder()
	app.handleUpload(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}
