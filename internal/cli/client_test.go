package cli

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"ok"}`))
	}))
	defer srv.Close()

	app := &App{Config: &Config{APIURL: srv.URL, AdminPassword: "test"}}
	c, err := app.NewClient()
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestCheckResponse(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		err := CheckResponse(http.StatusOK, nil)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		err := CheckResponse(http.StatusInternalServerError, []byte("bad"))
		assert.Error(t, err)
	})
}
