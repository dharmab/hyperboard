package cli

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestCheckResponse(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		if err := CheckResponse(http.StatusOK, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		if err := CheckResponse(http.StatusInternalServerError, []byte("bad")); err == nil {
			t.Error("expected error for 500 status")
		}
	})
}
