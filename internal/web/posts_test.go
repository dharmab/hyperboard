package web

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dharmab/hyperboard/pkg/client"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/gofrs/uuid/v5"
)

func TestHandlePosts(t *testing.T) {
	t.Parallel()
	postID := types.ID(uuid.Must(uuid.NewV4()))
	now := time.Now().UTC()
	posts := []types.Post{{ID: postID, MimeType: "image/webp", CreatedAt: now, UpdatedAt: now}}

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == postsAPIPath {
			jsonResponse(w, http.StatusOK, client.PostsResponse{Items: &posts})
			return
		}
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	app.handlePosts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), postID.String()) {
		t.Error("expected post ID in response body")
	}
}

func TestHandlePost_GET(t *testing.T) {
	t.Parallel()
	postID := types.ID(uuid.Must(uuid.NewV4()))
	now := time.Now().UTC()
	post := types.Post{
		ID:           postID,
		MimeType:     "image/webp",
		ContentUrl:   "http://storage/posts/" + postID.String() + "/content.webp",
		ThumbnailUrl: "http://storage/posts/" + postID.String() + "/thumbnail.webp",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/posts/") && strings.HasSuffix(r.URL.Path, "/similar") {
			jsonResponse(w, http.StatusOK, client.PostsResponse{})
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/v1/posts/") {
			jsonResponse(w, http.StatusOK, post)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/media/") {
			w.Header().Set("Content-Length", "1024")
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/posts/"+postID.String(), nil)
	req.SetPathValue("id", postID.String())
	w := httptest.NewRecorder()
	app.handlePost(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestHandlePost_DELETE(t *testing.T) {
	t.Parallel()
	postID := types.ID(uuid.Must(uuid.NewV4()))

	deleteCalled := false
	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/v1/posts/") {
			deleteCalled = true
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/posts/"+postID.String(), nil)
	req.SetPathValue("id", postID.String())
	w := httptest.NewRecorder()
	app.handlePost(w, req)

	if !deleteCalled {
		t.Error("expected delete to be called")
	}
	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusSeeOther)
	}
}

func TestHandlePost_InvalidID(t *testing.T) {
	t.Parallel()
	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/posts/not-a-uuid", nil)
	req.SetPathValue("id", "not-a-uuid")
	w := httptest.NewRecorder()
	app.handlePost(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (template render with error)", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), "Invalid post ID") {
		t.Error("expected error message about invalid ID")
	}
}

func TestHandlePosts_WithTagFilters(t *testing.T) {
	t.Parallel()
	posts := []types.Post{}

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == postsAPIPath {
			jsonResponse(w, http.StatusOK, client.PostsResponse{Items: &posts})
			return
		}
		http.NotFound(w, r)
	}))
	app.cfg.TagFilters = []tagFilter{{Label: "Rating", Tags: []string{"rating:safe", "rating:explicit"}}}

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	app.handlePosts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "tag-filter-btn") {
		t.Error("expected tag-filter-btn class in response body")
	}
	if !strings.Contains(body, "Rating") {
		t.Error("expected button label 'Rating' in response body")
	}
	if !strings.Contains(body, "data-tags") {
		t.Error("expected data-tags attribute in response body")
	}
}

func TestHandlePosts_WithoutTagFilters(t *testing.T) {
	t.Parallel()
	posts := []types.Post{}

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == postsAPIPath {
			jsonResponse(w, http.StatusOK, client.PostsResponse{Items: &posts})
			return
		}
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	app.handlePosts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	if strings.Contains(w.Body.String(), `class="tag-filters"`) {
		t.Error("expected no tag-filters div when TagFilters config is empty")
	}
}

func TestHandleTagSuggestions(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	tags := []types.Tag{
		{Name: "alpha", CreatedAt: now, UpdatedAt: now},
		{Name: "beta", CreatedAt: now, UpdatedAt: now},
		{Name: "gamma", CreatedAt: now, UpdatedAt: now},
	}

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/tags") {
			jsonResponse(w, http.StatusOK, client.TagsResponse{Items: &tags})
			return
		}
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/tag-suggestions", nil)
	w := httptest.NewRecorder()
	app.handleTagSuggestions(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	for _, name := range []string{"alpha", "beta", "gamma"} {
		if !strings.Contains(body, name) {
			t.Errorf("expected %q in response body", name)
		}
	}
}

func TestHandleTagSuggestions_Pagination(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()

	const totalTags = 3500
	const pageSize = 1000

	// Generate all tags
	allTags := make([]types.Tag, totalTags)
	for i := range allTags {
		allTags[i] = types.Tag{Name: fmt.Sprintf("tag-%04d", i), CreatedAt: now, UpdatedAt: now}
	}

	// Split into pages
	var pages [][]types.Tag
	for i := 0; i < len(allTags); i += pageSize {
		end := min(i+pageSize, len(allTags))
		pages = append(pages, allTags[i:end])
	}

	callCount := 0
	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/tags") {
			callCount++
			cursorParam := r.URL.Query().Get("cursor")
			pageIdx := 0
			if cursorParam != "" {
				_, _ = fmt.Sscanf(cursorParam, "page%d", &pageIdx)
			}
			resp := client.TagsResponse{Items: &pages[pageIdx]}
			if pageIdx+1 < len(pages) {
				next := fmt.Sprintf("page%d", pageIdx+1)
				resp.Cursor = &next
			}
			jsonResponse(w, http.StatusOK, resp)
			return
		}
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/tag-suggestions", nil)
	w := httptest.NewRecorder()
	app.handleTagSuggestions(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	for _, tag := range allTags {
		if !strings.Contains(body, fmt.Sprintf("value=%q", tag.Name)) {
			t.Errorf("expected %q in response body", tag.Name)
		}
	}
	expectedCalls := len(pages)
	if callCount != expectedCalls {
		t.Errorf("expected %d API calls, got %d", expectedCalls, callCount)
	}
}

func TestHandleTagSuggestions_FilterByQuery(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	tags := []types.Tag{
		{Name: "alpha", CreatedAt: now, UpdatedAt: now},
		{Name: "beta", CreatedAt: now, UpdatedAt: now},
		{Name: "gamma", CreatedAt: now, UpdatedAt: now},
	}

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/tags") {
			jsonResponse(w, http.StatusOK, client.TagsResponse{Items: &tags})
			return
		}
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/tag-suggestions?q=alph", nil)
	w := httptest.NewRecorder()
	app.handleTagSuggestions(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !strings.Contains(body, "alpha") {
		t.Error("expected alpha in response")
	}
	if strings.Contains(body, "beta") {
		t.Error("expected beta to be filtered out")
	}
}

func TestHandleTagSuggestions_ExcludeTags(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	tags := []types.Tag{
		{Name: "alpha", CreatedAt: now, UpdatedAt: now},
		{Name: "beta", CreatedAt: now, UpdatedAt: now},
	}

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/tags") {
			jsonResponse(w, http.StatusOK, client.TagsResponse{Items: &tags})
			return
		}
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/tag-suggestions?exclude=alpha", nil)
	w := httptest.NewRecorder()
	app.handleTagSuggestions(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if strings.Contains(body, "alpha") {
		t.Error("expected alpha to be excluded")
	}
	if !strings.Contains(body, "beta") {
		t.Error("expected beta in response")
	}
}
