package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dharmab/hyperboard/internal/types"
	"github.com/gofrs/uuid/v5"
)

type mockAPIClient struct {
	getFunc        func(ctx context.Context, path string, out any) error
	getWithQueryFn func(ctx context.Context, path string, query url.Values, out any) error
	getRawFn       func(ctx context.Context, path string) (*http.Response, error)
	headFn         func(ctx context.Context, path string) (*http.Response, error)
	postFn         func(ctx context.Context, path string, body any, out any) (int, error)
	putFn          func(ctx context.Context, path string, body any, out any) (int, error)
	deleteFn       func(ctx context.Context, path string) (int, error)
	uploadFileFn   func(ctx context.Context, data []byte, contentType string, force bool, out any) (int, []byte, error)
}

func (m *mockAPIClient) get(ctx context.Context, path string, out any) error {
	if m.getFunc != nil {
		return m.getFunc(ctx, path, out)
	}
	return errors.New("not implemented")
}

func (m *mockAPIClient) getWithQuery(ctx context.Context, path string, query url.Values, out any) error {
	if m.getWithQueryFn != nil {
		return m.getWithQueryFn(ctx, path, query, out)
	}
	return errors.New("not implemented")
}

func (m *mockAPIClient) getRaw(ctx context.Context, path string) (*http.Response, error) {
	if m.getRawFn != nil {
		return m.getRawFn(ctx, path)
	}
	return nil, errors.New("not implemented")
}

func (m *mockAPIClient) head(ctx context.Context, path string) (*http.Response, error) {
	if m.headFn != nil {
		return m.headFn(ctx, path)
	}
	return nil, errors.New("not implemented")
}

func (m *mockAPIClient) post(ctx context.Context, path string, body any, out any) (int, error) {
	if m.postFn != nil {
		return m.postFn(ctx, path, body, out)
	}
	return 0, errors.New("not implemented")
}

func (m *mockAPIClient) put(ctx context.Context, path string, body any, out any) (int, error) {
	if m.putFn != nil {
		return m.putFn(ctx, path, body, out)
	}
	return 0, errors.New("not implemented")
}

func (m *mockAPIClient) delete(ctx context.Context, path string) (int, error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, path)
	}
	return 0, errors.New("not implemented")
}

func (m *mockAPIClient) uploadFile(ctx context.Context, data []byte, contentType string, force bool, out any) (int, []byte, error) {
	if m.uploadFileFn != nil {
		return m.uploadFileFn(ctx, data, contentType, force, out)
	}
	return 0, nil, errors.New("not implemented")
}

func newTestApp(mock *mockAPIClient) *App {
	tmpls, err := parseTemplates()
	if err != nil {
		panic(fmt.Sprintf("failed to parse templates: %v", err))
	}
	return &App{
		cfg:   &Config{},
		api:   mock,
		tmpls: tmpls,
	}
}

func TestHandlePosts(t *testing.T) {
	t.Parallel()
	postID := types.ID(uuid.Must(uuid.NewV4()))
	posts := []types.Post{{ID: postID, MimeType: "image/webp"}}

	mock := &mockAPIClient{
		getWithQueryFn: func(ctx context.Context, path string, query url.Values, out any) error {
			if strings.HasPrefix(path, "/api/v1/posts") {
				resp := out.(*postsResponse)
				resp.Items = &posts
				return nil
			}
			return fmt.Errorf("unexpected path: %s", path)
		},
	}
	app := newTestApp(mock)

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
	post := types.Post{
		ID:           postID,
		MimeType:     "image/webp",
		ContentUrl:   "http://storage/posts/" + postID.String() + "/content.webp",
		ThumbnailUrl: "http://storage/posts/" + postID.String() + "/thumbnail.webp",
	}

	mock := &mockAPIClient{
		getFunc: func(ctx context.Context, path string, out any) error {
			p := out.(*types.Post)
			*p = post
			return nil
		},
		headFn: func(ctx context.Context, path string) (*http.Response, error) {
			return &http.Response{
				StatusCode:    http.StatusOK,
				ContentLength: 1024,
				Body:          http.NoBody,
			}, nil
		},
		getWithQueryFn: func(ctx context.Context, path string, query url.Values, out any) error {
			return nil
		},
	}
	app := newTestApp(mock)

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
	mock := &mockAPIClient{
		deleteFn: func(ctx context.Context, path string) (int, error) {
			deleteCalled = true
			return http.StatusNoContent, nil
		},
	}
	app := newTestApp(mock)

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
	mock := &mockAPIClient{}
	app := newTestApp(mock)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/upload", nil)
	w := httptest.NewRecorder()
	app.handleUpload(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleTags(t *testing.T) {
	t.Parallel()
	tags := []types.Tag{{Name: "test-tag"}}

	mock := &mockAPIClient{
		getWithQueryFn: func(ctx context.Context, path string, query url.Values, out any) error {
			if strings.HasPrefix(path, "/api/v1/tags") {
				resp := out.(*tagsResponse)
				resp.Items = &tags
				return nil
			}
			if strings.HasPrefix(path, "/api/v1/tagCategories") {
				resp := out.(*tagCategoriesResponse)
				cats := []types.TagCategory{}
				resp.Items = &cats
				return nil
			}
			return fmt.Errorf("unexpected path: %s", path)
		},
	}
	app := newTestApp(mock)

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
	notes := []types.Note{{ID: noteID, Title: "Test Note"}}

	mock := &mockAPIClient{
		getFunc: func(ctx context.Context, path string, out any) error {
			resp := out.(*notesResponse)
			resp.Items = &notes
			return nil
		},
	}
	app := newTestApp(mock)

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

	mock := &mockAPIClient{
		getWithQueryFn: func(ctx context.Context, path string, query url.Values, out any) error {
			if strings.HasPrefix(path, "/api/v1/posts") {
				resp := out.(*postsResponse)
				resp.Items = &posts
				return nil
			}
			return fmt.Errorf("unexpected path: %s", path)
		},
	}
	app := newTestApp(mock)
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

	mock := &mockAPIClient{
		getWithQueryFn: func(ctx context.Context, path string, query url.Values, out any) error {
			if strings.HasPrefix(path, "/api/v1/posts") {
				resp := out.(*postsResponse)
				resp.Items = &posts
				return nil
			}
			return fmt.Errorf("unexpected path: %s", path)
		},
	}
	app := newTestApp(mock)

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

	mock := &mockAPIClient{
		postFn: func(ctx context.Context, path string, body any, out any) (int, error) {
			if out != nil {
				data, err := json.Marshal(types.Note{ID: createdID, Title: "New Note"})
				if err != nil {
					panic(err)
				}
				_ = json.Unmarshal(data, out)
			}
			return http.StatusCreated, nil
		},
	}
	app := newTestApp(mock)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/notes", nil)
	w := httptest.NewRecorder()
	app.handleNotes(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusSeeOther)
	}
	loc := w.Header().Get("Location")
	if !strings.Contains(loc, createdID.String()) {
		t.Errorf("redirect location = %q, want it to contain %s", loc, createdID)
	}
}
