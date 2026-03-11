package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/dharmab/hyperboard/internal/db/store"
	"github.com/dharmab/hyperboard/pkg/types"
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
		{"日本語", true}, //nolint:gosmopolitan // intentional: testing Unicode letter support
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

func TestTagCascadeCRUD(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	ctx := t.Context()

	suffix := uuid.Must(uuid.NewV4()).String()[:8]
	parentName := "cascade-crud-parent-" + suffix
	childName := "cascade-crud-child-" + suffix

	// Create child tag first.
	childBody := types.Tag{Name: childName}
	cb, err := json.Marshal(childBody)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	childReq := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/tags/"+childName, bytes.NewReader(cb))
	childReq.Header.Set("Content-Type", "application/json")
	childW := httptest.NewRecorder()
	srv.PutTag(childW, childReq, childName)
	if childW.Code != http.StatusCreated {
		t.Fatalf("PutTag child status = %d, want %d; body = %s", childW.Code, http.StatusCreated, childW.Body.String())
	}

	// Create parent tag with cascade to child.
	cascades := []string{childName}
	parentBody := types.Tag{Name: parentName, CascadingTags: &cascades}
	pb, err := json.Marshal(parentBody)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	parentReq := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/tags/"+parentName, bytes.NewReader(pb))
	parentReq.Header.Set("Content-Type", "application/json")
	parentW := httptest.NewRecorder()
	srv.PutTag(parentW, parentReq, parentName)
	if parentW.Code != http.StatusCreated {
		t.Fatalf("PutTag parent status = %d, want %d; body = %s", parentW.Code, http.StatusCreated, parentW.Body.String())
	}

	// Retrieve parent and verify cascade is set.
	getReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/tags/"+parentName, nil)
	getW := httptest.NewRecorder()
	srv.GetTag(getW, getReq, parentName)
	if getW.Code != http.StatusOK {
		t.Fatalf("GetTag status = %d, want %d", getW.Code, http.StatusOK)
	}
	var got types.Tag
	if err := json.NewDecoder(getW.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if got.CascadingTags == nil || len(*got.CascadingTags) != 1 || (*got.CascadingTags)[0] != childName {
		t.Errorf("CascadingTags = %v, want [%q]", got.CascadingTags, childName)
	}

	// Verify GetPostCascadingTags via a tagged post.
	// Don't use tagPost which would clear cascades by calling UpsertTag without CascadingTags.
	post := insertTestPost(t)
	_, err = testStore.UpdatePost(ctx, post.ID, "", []string{parentName}, time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to tag post: %v", err)
	}
	cascadingTags, err := testStore.GetPostCascadingTags(ctx, post.ID)
	if err != nil {
		t.Fatalf("GetPostCascadingTags failed: %v", err)
	}
	found := false
	for _, name := range cascadingTags {
		if name == childName {
			found = true
		}
	}
	if !found {
		t.Errorf("expected child tag %q in cascading tags for post; got %v", childName, cascadingTags)
	}
}

func TestConvertTagToAlias(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	suffix := uuid.Must(uuid.NewV4()).String()[:8]
	sourceName := "convert-source-" + suffix
	targetName := "convert-target-" + suffix

	now := time.Now().UTC()

	// Create source and target tags.
	_, _, err := testStore.UpsertTag(ctx, sourceName, store.TagInput{Name: sourceName}, now)
	if err != nil {
		t.Fatalf("failed to create source tag: %v", err)
	}
	_, _, err = testStore.UpsertTag(ctx, targetName, store.TagInput{Name: targetName}, now)
	if err != nil {
		t.Fatalf("failed to create target tag: %v", err)
	}

	// Tag two posts: one with source, one with target.
	sourcePost := insertTestPost(t)
	targetPost := insertTestPost(t)
	tagPost(t, sourcePost.ID, sourceName)
	tagPost(t, targetPost.ID, targetName)

	// Verify initial post counts.
	sourceCounts, err := testStore.GetTagPostCounts(ctx, []uuid.UUID{})
	_ = sourceCounts
	if err != nil {
		t.Fatalf("GetTagPostCounts failed: %v", err)
	}

	// Convert source tag to alias of target.
	result, err := testStore.ConvertTagToAlias(ctx, sourceName, targetName)
	if err != nil {
		t.Fatalf("ConvertTagToAlias failed: %v", err)
	}
	if result.Tag.Name != targetName {
		t.Errorf("result tag name = %q, want %q", result.Tag.Name, targetName)
	}

	// Source tag should no longer exist.
	_, err = testStore.GetTag(ctx, sourceName)
	if err == nil {
		t.Error("source tag should not exist after conversion")
	}

	// Source name should now resolve as an alias to target.
	resolved, err := testStore.ResolveAlias(ctx, sourceName)
	if err != nil {
		t.Fatalf("ResolveAlias failed: %v", err)
	}
	if resolved != targetName {
		t.Errorf("ResolveAlias(%q) = %q, want %q", sourceName, resolved, targetName)
	}

	// The post that was tagged with source should now be tagged with target.
	updatedPost, err := testStore.GetPost(ctx, sourcePost.ID)
	if err != nil {
		t.Fatalf("GetPost failed: %v", err)
	}
	foundTarget := false
	for _, tag := range updatedPost.Tags {
		if tag.Name == targetName {
			foundTarget = true
		}
	}
	if !foundTarget {
		t.Errorf("post originally tagged with source should now be tagged with target; tags = %v", updatedPost.Tags)
	}
}

func TestTagsPagination(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)

	suffix := uuid.Must(uuid.NewV4()).String()[:8]
	for i := range 3 {
		name := fmt.Sprintf("pagination-tag-%d-%s", i, suffix)
		body := types.Tag{Name: name}
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("json.Marshal: %v", err)
		}
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/tags/"+name, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.PutTag(w, req, name)
		if w.Code != http.StatusCreated {
			t.Fatalf("PutTag status = %d, want %d", w.Code, http.StatusCreated)
		}
	}

	limit := 1
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/tags?limit=1", nil)
	w := httptest.NewRecorder()
	srv.GetTags(w, req, GetTagsParams{Limit: &limit})

	if w.Code != http.StatusOK {
		t.Fatalf("GetTags status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp TagsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Items == nil || len(*resp.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(*resp.Items))
	}
	if resp.Cursor == nil {
		t.Error("expected cursor for next page when there are more tags")
	}
}
