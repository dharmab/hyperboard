package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleUpload_GET(t *testing.T) {
	t.Parallel()
	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/upload", nil)
	w := httptest.NewRecorder()
	app.handleUpload(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `<input type="file" id="file-input" name="files"`)
}
