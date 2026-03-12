package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleMedia(t *testing.T) {
	t.Parallel()
	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/media/") {
			w.Header().Set("Content-Type", "image/webp")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("image-data"))
			return
		}
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/media/posts/abc/content.webp", nil)
	w := httptest.NewRecorder()
	app.handleMedia(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); ct != "image/webp" {
		t.Errorf("Content-Type = %q, want image/webp", ct)
	}
	if !strings.Contains(w.Body.String(), "image-data") {
		t.Error("expected proxied body content")
	}
}
