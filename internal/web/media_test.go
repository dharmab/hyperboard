package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

			contentType := w.Header().Get("Content-Type")
			contentDisposition := w.Header().Get("Content-Disposition")
			body := w.Body.String()
			require.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "image/webp", contentType)
			assert.Equal(t, tt.disposition, contentDisposition)
			assert.Equal(t, "image-data", body)
		})
	}
}

func TestProxyPostMediaRange(t *testing.T) {
	t.Parallel()
	const (
		postID    = "7a4dd550-2f65-4ab5-8c91-b11ee924370c"
		byteRange = "bytes=10-13"
	)

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/posts/"+postID+"/content", r.URL.Path)
		assert.Equal(t, byteRange, r.Header.Get("Range"))
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Length", "4")
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Range", "bytes 10-13/100")
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write([]byte("data"))
	}))

	mux := http.NewServeMux()
	app.registerRoutes(mux)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/posts/"+postID+"/content", nil)
	req.Header.Set("Range", byteRange)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	require.Equal(t, http.StatusPartialContent, w.Code)
	assert.Equal(t, "video/mp4", w.Header().Get("Content-Type"))
	assert.Equal(t, "4", w.Header().Get("Content-Length"))
	assert.Equal(t, "bytes", w.Header().Get("Accept-Ranges"))
	assert.Equal(t, "bytes 10-13/100", w.Header().Get("Content-Range"))
	assert.Equal(t, "data", w.Body.String())
}

func TestProxyPostMediaUnsatisfiableRange(t *testing.T) {
	t.Parallel()
	const postID = "7a4dd550-2f65-4ab5-8c91-b11ee924370c"

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "bytes=100-", r.Header.Get("Range"))
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Range", "bytes */100")
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		_, _ = w.Write([]byte(`{"message":"Invalid or unsatisfiable byte range"}`))
	}))

	mux := http.NewServeMux()
	app.registerRoutes(mux)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/posts/"+postID+"/content", nil)
	req.Header.Set("Range", "bytes=100-")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	require.Equal(t, http.StatusRequestedRangeNotSatisfiable, w.Code)
	assert.Equal(t, "bytes", w.Header().Get("Accept-Ranges"))
	assert.Equal(t, "bytes */100", w.Header().Get("Content-Range"))
	assert.Contains(t, w.Body.String(), "Invalid or unsatisfiable byte range")
}
