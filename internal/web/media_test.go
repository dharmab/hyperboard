package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxyPostMedia(t *testing.T) {
	t.Parallel()
	const postID = "7a4dd550-2f65-4ab5-8c91-b11ee924370c"

	tests := []struct {
		name        string
		path        string
		apiPath     string
		disposition string
	}{
		{
			name:    "content",
			path:    "/posts/" + postID + "/content",
			apiPath: "/api/v1/posts/" + postID + "/content",
		},
		{
			name:        "download",
			path:        "/posts/" + postID + "/download",
			apiPath:     "/api/v1/posts/" + postID + "/content/download",
			disposition: `attachment; filename="post.webp"`,
		},
		{
			name:    "thumbnail",
			path:    "/posts/" + postID + "/thumbnail",
			apiPath: "/api/v1/posts/" + postID + "/thumbnail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.apiPath {
					http.NotFound(w, r)
					return
				}
				w.Header().Set("Content-Type", "image/webp")
				w.Header().Set("Cache-Control", "private, max-age=86400")
				if tt.disposition != "" {
					w.Header().Set("Content-Disposition", tt.disposition)
				}
				_, _ = w.Write([]byte("image-data"))
			}))

			mux := http.NewServeMux()
			app.registerRoutes(mux)
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
			}
			if got := w.Header().Get("Content-Type"); got != "image/webp" {
				t.Errorf("Content-Type = %q, want image/webp", got)
			}
			if got := w.Header().Get("Content-Disposition"); got != tt.disposition {
				t.Errorf("Content-Disposition = %q, want %q", got, tt.disposition)
			}
			if got := w.Body.String(); got != "image-data" {
				t.Errorf("body = %q, want image-data", got)
			}
		})
	}
}
