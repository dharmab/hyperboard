package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/dharmab/hyperboard/internal/types"
	"github.com/gofrs/uuid/v5"
)

func TestIsValidName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		valid bool
	}{
		{"abc", true},
		{"ABC", true},
		{"123", true},
		{"1abc", true},
		{"café", true},
		{"日本語", true},
		{"", false},
		{"-abc", false},
		{"_abc", false},
		{" abc", false},
		{"!abc", false},
		{"abc ", false},
		{"abc  def", false},
		{"abc def", true},
		{"abc\t\tdef", false},
		{"abc\tdef", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isValidName(tc.name); got != tc.valid {
				t.Errorf("isValidName(%q) = %v, want %v", tc.name, got, tc.valid)
			}
		})
	}
}

func TestPutTagValidation(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)

	for _, name := range []string{"-bad", "_bad", " bad", "!bad", "bad ", "bad  tag"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			body := types.Tag{Name: name, Description: "test"}
			b, _ := json.Marshal(body)
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/tags/"+url.PathEscape(name), bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			srv.PutTag(w, req, name)
			if w.Code != http.StatusBadRequest {
				t.Errorf("PutTag(%q) status = %d, want %d", name, w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestTagsIntegration(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)

	tagName := "test-tag-" + uuid.Must(uuid.NewV4()).String()[:8]

	t.Run("create tag", func(t *testing.T) {
		body := types.Tag{
			Name:        tagName,
			Description: "A test tag",
		}
		b, _ := json.Marshal(body)
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/tags/"+tagName, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.PutTag(w, req, tagName)

		if w.Code != http.StatusCreated {
			t.Fatalf("PutTag create status = %d, want %d; body = %s", w.Code, http.StatusCreated, w.Body.String())
		}

		var tag types.Tag
		if err := json.NewDecoder(w.Body).Decode(&tag); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if tag.Name != tagName {
			t.Errorf("Name = %q, want %q", tag.Name, tagName)
		}
	})

	t.Run("get tag", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/tags/"+tagName, nil)
		w := httptest.NewRecorder()
		srv.GetTag(w, req, tagName)

		if w.Code != http.StatusOK {
			t.Fatalf("GetTag status = %d, want %d", w.Code, http.StatusOK)
		}

		var tag types.Tag
		if err := json.NewDecoder(w.Body).Decode(&tag); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if tag.Name != tagName {
			t.Errorf("Name = %q, want %q", tag.Name, tagName)
		}
	})

	t.Run("list tags", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/tags", nil)
		w := httptest.NewRecorder()
		srv.GetTags(w, req, GetTagsParams{})

		if w.Code != http.StatusOK {
			t.Fatalf("GetTags status = %d, want %d", w.Code, http.StatusOK)
		}

		var resp TagsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Items == nil || len(*resp.Items) == 0 {
			t.Error("expected at least one tag")
		}
	})

	t.Run("update tag description", func(t *testing.T) {
		body := types.Tag{
			Name:        tagName,
			Description: "Updated description",
		}
		b, _ := json.Marshal(body)
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/tags/"+tagName, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.PutTag(w, req, tagName)

		if w.Code != http.StatusOK {
			t.Fatalf("PutTag update status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
		}

		var tag types.Tag
		if err := json.NewDecoder(w.Body).Decode(&tag); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if tag.Description != "Updated description" {
			t.Errorf("Description = %q, want %q", tag.Description, "Updated description")
		}
	})

	t.Run("rename tag", func(t *testing.T) {
		newName := tagName + "-renamed"
		body := types.Tag{
			Name:        newName,
			Description: "Updated description",
		}
		b, _ := json.Marshal(body)
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/tags/"+tagName, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.PutTag(w, req, tagName)

		if w.Code != http.StatusOK {
			t.Fatalf("PutTag rename status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
		}

		// Update tagName for subsequent tests
		tagName = newName
	})

	t.Run("set tag aliases", func(t *testing.T) {
		aliases := []string{"alias1", "alias2"}
		body := types.Tag{
			Name:    tagName,
			Aliases: &aliases,
		}
		b, _ := json.Marshal(body)
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/tags/"+tagName, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.PutTag(w, req, tagName)

		if w.Code != http.StatusOK {
			t.Fatalf("PutTag aliases status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
		}

		var tag types.Tag
		if err := json.NewDecoder(w.Body).Decode(&tag); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if tag.Aliases == nil || len(*tag.Aliases) != 2 {
			t.Errorf("expected 2 aliases, got %v", tag.Aliases)
		}
	})

	t.Run("assign tag category", func(t *testing.T) {
		// Create a category first
		catName := "tag-test-cat-" + uuid.Must(uuid.NewV4()).String()[:8]
		catBody := types.TagCategory{Name: catName, Description: "For tag test"}
		cb, _ := json.Marshal(catBody)
		catReq := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/tagCategories/"+catName, bytes.NewReader(cb))
		catReq.Header.Set("Content-Type", "application/json")
		catW := httptest.NewRecorder()
		srv.PutTagCategory(catW, catReq, catName)
		if catW.Code != http.StatusCreated {
			t.Fatalf("PutTagCategory status = %d, want %d", catW.Code, http.StatusCreated)
		}

		// Assign category to tag
		body := types.Tag{
			Name:     tagName,
			Category: &catName,
		}
		b, _ := json.Marshal(body)
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/tags/"+tagName, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.PutTag(w, req, tagName)

		if w.Code != http.StatusOK {
			t.Fatalf("PutTag category status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
		}

		var tag types.Tag
		if err := json.NewDecoder(w.Body).Decode(&tag); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if tag.Category == nil || *tag.Category != catName {
			t.Errorf("Category = %v, want %q", tag.Category, catName)
		}
	})

	t.Run("delete tag", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/api/v1/tags/"+tagName, nil)
		w := httptest.NewRecorder()
		srv.DeleteTag(w, req, tagName)

		if w.Code != http.StatusNoContent {
			t.Fatalf("DeleteTag status = %d, want %d", w.Code, http.StatusNoContent)
		}
	})

	t.Run("get deleted tag returns not found", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/tags/"+tagName, nil)
		w := httptest.NewRecorder()
		srv.GetTag(w, req, tagName)

		if w.Code != http.StatusNotFound {
			t.Fatalf("GetTag status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}
