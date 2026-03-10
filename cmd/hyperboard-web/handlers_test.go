package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dharmab/hyperboard/pkg/client"
	"github.com/dharmab/hyperboard/internal/types"
	"github.com/gofrs/uuid/v5"
)

func newTestApp(t *testing.T, handler http.Handler) *App {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	tmpls, err := parseTemplates()
	if err != nil {
		t.Fatalf("failed to parse templates: %v", err)
	}

	api, err := newAPIClient(srv.URL, "test")
	if err != nil {
		t.Fatalf("failed to create API client: %v", err)
	}

	return &App{
		cfg:   &Config{},
		api:   api,
		media: newMediaClient(srv.URL, "test"),
		tmpls: tmpls,
	}
}

const postsAPIPath = "/api/v1/posts"

func jsonResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	data, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(data)
}

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

func TestHandleUpload_GET(t *testing.T) {
	t.Parallel()
	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/upload", nil)
	w := httptest.NewRecorder()
	app.handleUpload(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleTags(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	tags := []types.Tag{{Name: "test-tag", CreatedAt: now, UpdatedAt: now}}
	cats := []types.TagCategory{}

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/tags") {
			jsonResponse(w, http.StatusOK, client.TagsResponse{Items: &tags})
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/v1/tagCategories") {
			jsonResponse(w, http.StatusOK, client.TagCategoriesResponse{Items: &cats})
			return
		}
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/tags", nil)
	w := httptest.NewRecorder()
	app.handleTags(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "test-tag") {
		t.Error("expected tag name in response body")
	}
}

func TestHandleNotes_GET(t *testing.T) {
	t.Parallel()
	noteID := types.ID(uuid.Must(uuid.NewV4()))
	now := time.Now().UTC()
	notes := []types.Note{{ID: noteID, Title: "Test Note", CreatedAt: now, UpdatedAt: now}}

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/notes" {
			jsonResponse(w, http.StatusOK, client.NotesResponse{Items: &notes})
			return
		}
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/notes", nil)
	w := httptest.NewRecorder()
	app.handleNotes(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Test Note") {
		t.Error("expected note title in response body")
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
	app.cfg.TagFilters = `[{"label":"Rating","tags":["rating:safe","rating:explicit"]}]`

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

func TestHandleNotes_POST(t *testing.T) {
	t.Parallel()
	createdID := types.ID(uuid.Must(uuid.NewV4()))
	now := time.Now().UTC()

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v1/notes" {
			jsonResponse(w, http.StatusCreated, types.Note{ID: createdID, Title: "New Note", CreatedAt: now, UpdatedAt: now})
			return
		}
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/notes", nil)
	w := httptest.NewRecorder()
	app.handleNotes(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusSeeOther, w.Body.String())
	}
	loc := w.Header().Get("Location")
	if !strings.Contains(loc, createdID.String()) {
		t.Errorf("redirect location = %q, want it to contain %s", loc, createdID)
	}
}

func TestHandleMedia(t *testing.T) {
	t.Parallel()
	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/media/") {
			w.Header().Set("Content-Type", "image/webp")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("image-data"))
			return
		}
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/media/posts/abc/content.webp", nil)
	w := httptest.NewRecorder()
	app.handleMedia(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); ct != "image/webp" {
		t.Errorf("Content-Type = %q, want image/webp", ct)
	}
	if !strings.Contains(w.Body.String(), "image-data") {
		t.Error("expected proxied body content")
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

func TestHandleTagConvertToAlias(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/convert-to-alias") {
			jsonResponse(w, http.StatusOK, types.Tag{Name: "target-tag", CreatedAt: now, UpdatedAt: now})
			return
		}
		http.NotFound(w, r)
	}))

	form := strings.NewReader("target=target-tag")
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/tags/source-tag/convert-to-alias", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("name", "source-tag")
	w := httptest.NewRecorder()
	app.handleTagConvertToAlias(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusSeeOther, w.Body.String())
	}
	loc := w.Header().Get("Location")
	if loc != "/tags/target-tag" {
		t.Errorf("redirect location = %q, want /tags/target-tag", loc)
	}
}

func TestHandleTagConvertToAlias_EmptyTarget(t *testing.T) {
	t.Parallel()
	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/tags/source-tag/convert-to-alias", nil)
	req.SetPathValue("name", "source-tag")
	w := httptest.NewRecorder()
	app.handleTagConvertToAlias(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleTagConvertToAlias_SameTarget(t *testing.T) {
	t.Parallel()
	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))

	form := strings.NewReader("target=source-tag")
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/tags/source-tag/convert-to-alias", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("name", "source-tag")
	w := httptest.NewRecorder()
	app.handleTagConvertToAlias(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}
