package cli

import (
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestCheckResponseStatus(t *testing.T) {
	t.Parallel()
	require.NoError(t, CheckResponseStatus(http.StatusOK, nil, http.StatusOK))
	require.ErrorContains(t, CheckResponseStatus(http.StatusCreated, nil, http.StatusOK), "expected HTTP 200, got HTTP 201")
	require.ErrorContains(t, CheckResponseStatus(http.StatusInternalServerError, []byte("bad"), http.StatusOK), "bad")
}

func TestCheckJSONResponse(t *testing.T) {
	t.Parallel()
	payload := struct{}{}
	require.NoError(t, CheckJSONResponse(http.StatusOK, nil, http.StatusOK, &payload))
	require.ErrorContains(t, CheckJSONResponse[struct{}](http.StatusOK, nil, http.StatusOK, nil), "did not contain the expected JSON body")
}
