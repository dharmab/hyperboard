package tagcategories

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/dharmab/hyperboard/internal/cli"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTagCategoryRefusesExistingCategory(t *testing.T) {
	t.Parallel()
	var putRequests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
		case http.MethodPut:
			putRequests.Add(1)
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	app := &cli.App{Config: &cli.Config{APIURL: srv.URL, AdminPassword: "test"}}
	err := createTagCategory(app, "existing", types.TagCategory{Name: "existing", Color: "#888888"})
	require.ErrorContains(t, err, "tagcategory/existing already exists")
	assert.Zero(t, putRequests.Load(), "create must not call the upsert endpoint for an existing tag category")
}

func TestCreateTagCategoryHandlesPutUpdateResponse(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		case http.MethodPut:
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	app := &cli.App{Config: &cli.Config{APIURL: srv.URL, AdminPassword: "test"}}
	err := createTagCategory(app, "raced", types.TagCategory{Name: "raced", Color: "#888888"})
	require.ErrorContains(t, err, "tagcategory/raced already exists")
	require.ErrorContains(t, err, "update response")
}
