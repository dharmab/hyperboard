package posts

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/dharmab/hyperboard/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTagCSV(t *testing.T) {
	t.Parallel()

	tags, err := parseTagCSV(" landscape,night sky , portrait ")
	require.NoError(t, err)
	assert.Equal(t, []string{"landscape", "night sky", "portrait"}, tags)

	for _, value := range []string{" ", ",portrait", "landscape,", "landscape,,portrait", "-invalid", "two  spaces"} {
		t.Run(value, func(t *testing.T) {
			t.Parallel()
			_, err := parseTagCSV(value)
			assert.Error(t, err)
		})
	}
}

func TestCreatePostValidatesTagsBeforeReadingOrUploading(t *testing.T) {
	t.Parallel()

	err := createPost(&cli.App{}, filepath.Join(t.TempDir(), "missing.png"), "landscape, ", "")
	require.ErrorContains(t, err, "tag 2 is empty")
	assert.NotContains(t, err.Error(), "read ")
}

func TestCreatePostMetadataFailureIncludesCreatedPostID(t *testing.T) {
	t.Parallel()
	const postID = "123e4567-e89b-12d3-a456-426614174000"
	var uploadRequests atomic.Int32
	var putRequests atomic.Int32
	var deleteRequests atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			uploadRequests.Add(1)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = fmt.Fprintf(w, `{"post":{"id":"%s","mimeType":"image/png","contentUrl":"/content","thumbnailUrl":"/thumbnail","note":"","hasAudio":false,"tags":[],"createdAt":"2024-01-01T00:00:00Z","updatedAt":"2024-01-01T00:00:00Z"}}`, postID)
		case http.MethodPut:
			putRequests.Add(1)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message":"invalid metadata"}`))
		case http.MethodDelete:
			deleteRequests.Add(1)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	png, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=")
	require.NoError(t, err)
	filePath := filepath.Join(t.TempDir(), "post.png")
	require.NoError(t, os.WriteFile(filePath, png, 0o600))

	app := &cli.App{Config: &cli.Config{APIURL: srv.URL, AdminPassword: "test"}}
	err = createPost(app, filePath, " landscape, portrait ", "test note")
	require.Error(t, err)
	assert.Contains(t, err.Error(), postID)
	assert.Contains(t, err.Error(), "was created, but updating its metadata failed")
	assert.Contains(t, err.Error(), "hyperboardctl edit post "+postID)
	assert.EqualValues(t, 1, uploadRequests.Load())
	assert.EqualValues(t, 1, putRequests.Load())
	assert.Zero(t, deleteRequests.Load(), "metadata failure must not trigger destructive cleanup")
}
