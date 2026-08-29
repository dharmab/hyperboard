package replace

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"uuid"

	"github.com/dharmab/hyperboard/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplaceMediaUsesDetectedContentType(t *testing.T) {
	t.Parallel()
	postID := uuid.New().String()
	pngData := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")
	filePath := filepath.Join(t.TempDir(), "media.png")
	require.NoError(t, os.WriteFile(filePath, pngData, 0o600))

	for _, tt := range []struct {
		name string
		path string
		run  func(*cli.App) error
	}{
		{
			name: "content",
			path: "/api/v1/posts/" + postID + "/content",
			run: func(app *cli.App) error {
				return replaceContent(app, postID, filePath)
			},
		},
		{
			name: "thumbnail",
			path: "/api/v1/posts/" + postID + "/thumbnail",
			run: func(app *cli.App) error {
				return replaceThumbnail(app, postID, filePath)
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var contentType string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Equal(t, tt.path, r.URL.Path)
				contentType = r.Header.Get("Content-Type")
				http.Error(w, "stop after request inspection", http.StatusTeapot)
			}))
			t.Cleanup(server.Close)

			app := &cli.App{Config: &cli.Config{APIURL: server.URL, AdminPassword: "test"}}
			err := tt.run(app)
			require.Error(t, err)
			assert.Equal(t, "image/png", contentType)
		})
	}
}
