package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dharmab/hyperboard/internal/types"
	"github.com/gofrs/uuid/v5"
)

func TestTagCategoriesIntegration(t *testing.T) {
	srv := newTestServer(t)

	catName := "test-category-" + uuid.Must(uuid.NewV4()).String()[:8]

	t.Run("create tag category", func(t *testing.T) {
		body := types.TagCategory{
			Name:        catName,
			Description: "A test category",
			Color:       "#ff0000",
		}
		b, _ := json.Marshal(body)
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/tagCategories/"+catName, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.PutTagCategory(w, req, catName)

		if w.Code != http.StatusCreated {
			t.Fatalf("PutTagCategory status = %d, want %d; body = %s", w.Code, http.StatusCreated, w.Body.String())
		}

		var cat types.TagCategory
		if err := json.NewDecoder(w.Body).Decode(&cat); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if cat.Name != catName {
			t.Errorf("Name = %q, want %q", cat.Name, catName)
		}
	})

	t.Run("get tag category", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/tagCategories/"+catName, nil)
		w := httptest.NewRecorder()
		srv.GetTagCategory(w, req, catName)

		if w.Code != http.StatusOK {
			t.Fatalf("GetTagCategory status = %d, want %d", w.Code, http.StatusOK)
		}

		var cat types.TagCategory
		if err := json.NewDecoder(w.Body).Decode(&cat); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if cat.Name != catName {
			t.Errorf("Name = %q, want %q", cat.Name, catName)
		}
	})

	t.Run("list tag categories", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/tagCategories", nil)
		w := httptest.NewRecorder()
		srv.GetTagCategories(w, req, GetTagCategoriesParams{})

		if w.Code != http.StatusOK {
			t.Fatalf("GetTagCategories status = %d, want %d", w.Code, http.StatusOK)
		}

		var resp TagCategoriesResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Items == nil || len(*resp.Items) == 0 {
			t.Error("expected at least one tag category")
		}
	})

	t.Run("update tag category", func(t *testing.T) {
		body := types.TagCategory{
			Name:        catName,
			Description: "Updated description",
			Color:       "#00ff00",
		}
		b, _ := json.Marshal(body)
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/tagCategories/"+catName, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.PutTagCategory(w, req, catName)

		if w.Code != http.StatusOK {
			t.Fatalf("PutTagCategory update status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
		}

		var cat types.TagCategory
		if err := json.NewDecoder(w.Body).Decode(&cat); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if cat.Description != "Updated description" {
			t.Errorf("Description = %q, want %q", cat.Description, "Updated description")
		}
	})

	t.Run("delete tag category", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/api/v1/tagCategories/"+catName, nil)
		w := httptest.NewRecorder()
		srv.DeleteTagCategory(w, req, catName)

		if w.Code != http.StatusNoContent {
			t.Fatalf("DeleteTagCategory status = %d, want %d", w.Code, http.StatusNoContent)
		}
	})

	t.Run("get deleted tag category returns not found", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/tagCategories/"+catName, nil)
		w := httptest.NewRecorder()
		srv.GetTagCategory(w, req, catName)

		if w.Code != http.StatusNotFound {
			t.Fatalf("GetTagCategory status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}
