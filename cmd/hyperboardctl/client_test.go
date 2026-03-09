package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/gofrs/uuid/v5"
)

func TestFetchAll(t *testing.T) {
	t.Parallel()
	postID := types.ID(uuid.Must(uuid.NewV4()))
	now := time.Now().UTC()
	posts := []types.Post{{
		ID:        postID,
		MimeType:  "image/webp",
		CreatedAt: now,
		UpdatedAt: now,
	}}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/posts" {
			http.NotFound(w, r)
			return
		}
		resp := struct {
			Items *[]types.Post `json:"items"`
		}{Items: &posts}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	cfg := &Config{APIURL: srv.URL, AdminPassword: "test"}
	result, err := fetchAll[types.Post](cfg, srv.URL+"/api/v1/posts", nil)
	if err != nil {
		t.Fatalf("fetchAll error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 post, got %d", len(result))
	}
	if result[0].ID != postID {
		t.Errorf("ID = %v, want %v", result[0].ID, postID)
	}
}

func TestDoRequest(t *testing.T) {
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

	cfg := &Config{AdminPassword: "test"}
	resp, err := doRequest(cfg, http.MethodGet, srv.URL+"/api/v1/tags", "", nil)
	if err != nil {
		t.Fatalf("doRequest error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestCheckStatus(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		resp := &http.Response{StatusCode: 200, Body: http.NoBody}
		if err := checkStatus(resp); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		resp := &http.Response{StatusCode: 500, Body: http.NoBody}
		if err := checkStatus(resp); err == nil {
			t.Error("expected error for 500 status")
		}
	})
}
