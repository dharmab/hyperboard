package tags

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dharmab/hyperboard/internal/cli"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchAllTags(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	tags := []types.Tag{{
		Name:      "test-tag",
		CreatedAt: now,
		UpdatedAt: now,
	}}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/tags" {
			http.NotFound(w, r)
			return
		}
		resp := struct {
			Items *[]types.Tag `json:"items"`
		}{Items: &tags}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	app := &cli.App{Config: &cli.Config{APIURL: srv.URL, AdminPassword: "test"}}
	c, err := app.NewClient()
	require.NoError(t, err)
	result, err := FetchAllTags(c)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "test-tag", result[0].Name)
}
