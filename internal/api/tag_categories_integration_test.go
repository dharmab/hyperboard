package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/gofrs/uuid/v5"
)

func TestPutTagCategoryValidation(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)

	for _, name := range []string{"-bad", "_bad", " bad", "!bad", "bad ", "bad  category"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			body := types.TagCategory{Name: name, Description: "test", Color: "#ff0000"}
			b, _ := json.Marshal(body)
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/tagCategories/"+url.PathEscape(name), bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			srv.PutTagCategory(w, req, name)
			if w.Code != http.StatusBadRequest {
				t.Errorf("PutTagCategory(%q) status = %d, want %d", name, w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestTagCategoriesIntegration(t *testing.T) {
	t.Parallel()
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

func TestTagCategoriesPagination(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)

	suffix := uuid.Must(uuid.NewV4()).String()[:8]
	for i := range 3 {
		name := fmt.Sprintf("paginationcat-%d-%s", i, suffix)
		body := types.TagCategory{Name: name, Description: "pagination test", Color: "#000000"}
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("json.Marshal: %v", err)
		}
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/tagCategories/"+name, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.PutTagCategory(w, req, name)
		if w.Code != http.StatusCreated {
			t.Fatalf("PutTagCategory status = %d, want %d; body = %s", w.Code, http.StatusCreated, w.Body.String())
		}
	}

	limit := 1
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/tagCategories?limit=1", nil)
	w := httptest.NewRecorder()
	srv.GetTagCategories(w, req, GetTagCategoriesParams{Limit: &limit})

	if w.Code != http.StatusOK {
		t.Fatalf("GetTagCategories status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp TagCategoriesResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Items == nil || len(*resp.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(*resp.Items))
	}
	if resp.Cursor == nil {
		t.Error("expected cursor for next page when there are more categories")
	}
}
