package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dharmab/hyperboard/internal/types"
	"github.com/google/uuid"
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

	testCfg := &Config{APIURL: srv.URL, AdminPassword: "test"}
	c, err := newClient(testCfg)
	if err != nil {
		t.Fatalf("newClient error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestCheckResponse(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		if err := checkResponse(http.StatusOK, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		if err := checkResponse(http.StatusInternalServerError, []byte("bad")); err == nil {
			t.Error("expected error for 500 status")
		}
	})
}

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

	testCfg := &Config{APIURL: srv.URL, AdminPassword: "test"}
	c, err := newClient(testCfg)
	if err != nil {
		t.Fatalf("newClient error: %v", err)
	}
	result, err := fetchAllTags(c)
	if err != nil {
		t.Fatalf("fetchAllTags error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(result))
	}
	if result[0].Name != "test-tag" {
		t.Errorf("Name = %v, want test-tag", result[0].Name)
	}
}

func TestParseID(t *testing.T) {
	t.Parallel()
	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		expected := uuid.New()
		got, err := parseID(expected.String())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != expected {
			t.Errorf("got %v, want %v", got, expected)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()
		_, err := parseID("not-a-uuid")
		if err == nil {
			t.Error("expected error for invalid UUID")
		}
	})
}
