package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"uuid"

	"github.com/dharmab/hyperboard/pkg/client"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlePosts(t *testing.T) {
	t.Parallel()
	postID := types.ID(uuid.NewV4())
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

	body := w.Body.String()
	postIDString := postID.String()
	require.Equal(t, http.StatusOK, w.Code, "body = %s", body)
	assert.Contains(t, body, postIDString)
	assert.Contains(t, body, `hx-swap:inherited="innerMorph"`)
	assert.Contains(t, body, `id="post-`+postIDString+`"`)
}

func TestHandlePost_GET(t *testing.T) {
	t.Parallel()
	postID := types.ID(uuid.NewV4())
	now := time.Now().UTC()
	post := types.Post{
		ID:           postID,
		MimeType:     "video/mp4",
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
		if strings.HasPrefix(r.URL.Path, "/api/v1/posts/") && strings.HasSuffix(r.URL.Path, "/content") {
			w.Header().Set("Content-Length", "1024")
			w.WriteHeader(http.StatusOK)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/v1/posts/") {
			jsonResponse(w, http.StatusOK, post)
			return
		}
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/posts/"+postID.String(), nil)
	req.SetPathValue("id", postID.String())
	w := httptest.NewRecorder()
	app.handlePost(w, req)

	body := w.Body.String()
	assert.Equal(t, http.StatusOK, w.Code, "body = %s", body)
	assert.Contains(t, body, "hx-morph-skip")
	assert.Contains(t, body, `hx-swap="outerMorph"`)
}

func TestHandlePost_DELETE(t *testing.T) {
	t.Parallel()
	postID := types.ID(uuid.NewV4())

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

	redirect := w.Header().Get("HX-Redirect")
	assert.True(t, deleteCalled, "expected delete to be called")
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "/", redirect)
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

	body := w.Body.String()
	require.Equal(t, http.StatusOK, w.Code, "template render with error")
	assert.Contains(t, body, "Invalid post ID")
}

func TestHandlePostNote_GETDoesNotUpdate(t *testing.T) {
	t.Parallel()
	postID := types.ID(uuid.NewV4())
	apiCalled := false
	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
		http.Error(w, "unexpected API request", http.StatusInternalServerError)
	}))

	mux := http.NewServeMux()
	app.registerRoutes(mux)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/posts/"+postID.String()+"/note?note=replaced", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Equal(t, http.MethodPut, w.Header().Get("Allow"))
	assert.False(t, apiCalled)
}

func TestHandlePostNote_RejectsInvalidFormWithoutUpdating(t *testing.T) {
	t.Parallel()
	postID := types.ID(uuid.NewV4())

	tests := []struct {
		name   string
		body   string
		status int
	}{
		{name: "malformed", body: "note=%zz", status: http.StatusBadRequest},
		{name: "oversized", body: "note=" + strings.Repeat("x", int(maxFormBody)), status: http.StatusRequestEntityTooLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			apiCalled := false
			app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				apiCalled = true
				http.Error(w, "unexpected API request", http.StatusInternalServerError)
			}))

			mux := http.NewServeMux()
			app.registerRoutes(mux)
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/posts/"+postID.String()+"/note", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			assert.Equal(t, tt.status, w.Code)
			assert.False(t, apiCalled, "the existing post must not be fetched or updated after a form parse failure")
		})
	}
}

func TestHandlePostTags_TrimsTagName(t *testing.T) {
	t.Parallel()
	postID := types.ID(uuid.NewV4())
	now := time.Now().UTC()
	post := types.Post{
		ID:        postID,
		MimeType:  "image/webp",
		Note:      "",
		Tags:      []types.TagName{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/posts/"+postID.String() {
			http.NotFound(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			jsonResponse(w, http.StatusOK, post)
		case http.MethodPut:
			if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			jsonResponse(w, http.StatusOK, post)
		default:
			http.NotFound(w, r)
		}
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/posts/"+postID.String()+"/tags", strings.NewReader("q=example+"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", postID.String())
	w := httptest.NewRecorder()
	app.handlePostTags(w, req)

	body := w.Body.String()
	require.Equal(t, http.StatusOK, w.Code, "body = %s", body)
	assert.Equal(t, []types.TagName{"example"}, post.Tags)
	assert.Equal(t, "postTagsUpdated", w.Header().Get("HX-Trigger"))
}

func TestHandlePostTags_DuplicateIsNoOp(t *testing.T) {
	t.Parallel()
	postID := types.ID(uuid.NewV4())
	now := time.Now().UTC()
	post := types.Post{
		ID:        postID,
		MimeType:  "image/webp",
		Tags:      []types.TagName{"example"},
		CreatedAt: now,
		UpdatedAt: now,
	}
	putCalled := false

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/posts/"+postID.String() {
			http.NotFound(w, r)
			return
		}
		if r.Method == http.MethodPut {
			putCalled = true
			http.Error(w, "duplicate tag", http.StatusBadRequest)
			return
		}
		jsonResponse(w, http.StatusOK, post)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/posts/"+postID.String()+"/tags", strings.NewReader("q=example"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", postID.String())
	w := httptest.NewRecorder()
	app.handlePostTags(w, req)

	body := w.Body.String()
	require.Equal(t, http.StatusOK, w.Code, "body = %s", body)
	assert.False(t, putCalled)
	assert.Equal(t, []types.TagName{"example"}, post.Tags)
	assert.Equal(t, "postTagsUpdated", w.Header().Get("HX-Trigger"))
}

func TestHandlePostTags_ConcurrentDuplicateIsSuccess(t *testing.T) {
	t.Parallel()
	postID := types.ID(uuid.NewV4())
	now := time.Now().UTC()
	post := types.Post{
		ID:        postID,
		MimeType:  "image/webp",
		CreatedAt: now,
		UpdatedAt: now,
	}
	getCalls := 0

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/posts/"+postID.String() {
			http.NotFound(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			getCalls++
			if getCalls > 1 {
				post.Tags = []types.TagName{"example"}
			}
			jsonResponse(w, http.StatusOK, post)
		case http.MethodPut:
			http.Error(w, "conflict", http.StatusConflict)
		default:
			http.NotFound(w, r)
		}
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/posts/"+postID.String()+"/tags", strings.NewReader("q=example"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", postID.String())
	w := httptest.NewRecorder()
	app.handlePostTags(w, req)

	body := w.Body.String()
	require.Equal(t, http.StatusOK, w.Code, "body = %s", body)
	assert.Equal(t, 2, getCalls)
	assert.Contains(t, body, "example")
	assert.Equal(t, "postTagsUpdated", w.Header().Get("HX-Trigger"))
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

	body := w.Body.String()
	require.Equal(t, http.StatusOK, w.Code, "body = %s", body)
	assert.Contains(t, body, "tag-filter-btn")
	assert.Contains(t, body, "Rating")
	assert.Contains(t, body, "data-tags")
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

	body := w.Body.String()
	require.Equal(t, http.StatusOK, w.Code, "body = %s", body)
	assert.NotContains(t, body, `class="tag-filters"`)
}

func TestHandleTagSuggestions(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	pc1, pc2, pc3 := 10, 5, 1
	tags := []types.Tag{
		{Name: "alpha", PostCount: &pc1, CreatedAt: now, UpdatedAt: now},
		{Name: "apex", PostCount: &pc2, CreatedAt: now, UpdatedAt: now},
		{Name: "zulu", PostCount: &pc3, CreatedAt: now, UpdatedAt: now},
	}

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/tags") {
			jsonResponse(w, http.StatusOK, client.TagsResponse{Items: &tags})
			return
		}
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/tag-suggestions?q=a", nil)
	w := httptest.NewRecorder()
	app.handleTagSuggestions(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	for _, name := range []string{"alpha", "apex"} {
		expected := fmt.Sprintf(`data-value=%q`, name)
		assert.Contains(t, body, expected)
	}
	assert.NotContains(t, body, "zulu")
}

func TestHandleTagSuggestions_EscapesTagName(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	count := 1
	tags := []types.Tag{{
		Name:      `needle"><script>alert('x')</script>&`,
		PostCount: &count,
		CreatedAt: now,
		UpdatedAt: now,
	}}

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, http.StatusOK, client.TagsResponse{Items: &tags})
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/tag-suggestions?q=needle", nil)
	w := httptest.NewRecorder()
	app.handleTagSuggestions(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	escaped := "needle&#34;&gt;&lt;script&gt;alert(&#39;x&#39;)&lt;/script&gt;&amp;"
	assert.Contains(t, body, `data-value="`+escaped+`"`)
	assert.Contains(t, body, `>`+escaped+` <span class="ac-count">`)
	assert.NotContains(t, body, "<script>")
	assert.NotContains(t, body, `data-value="needle">`)
}

func TestHandleTagSuggestions_Pagination(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()

	const totalTags = 3500
	const pageSize = 1000

	// Generate all tags with "tag-" prefix so they all match q=tag
	allTags := make([]types.Tag, totalTags)
	for i := range allTags {
		pc := totalTags - i
		allTags[i] = types.Tag{Name: fmt.Sprintf("tag-%04d", i), PostCount: &pc, CreatedAt: now, UpdatedAt: now}
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

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/tag-suggestions?q=tag", nil)
	w := httptest.NewRecorder()
	app.handleTagSuggestions(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	// Results should be capped at 20
	resultCount := strings.Count(body, "ac-item")
	assert.Equal(t, 20, resultCount)
	assert.Len(t, pages, callCount)
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

	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, `data-value="alpha"`)
	assert.NotContains(t, body, "beta")
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

	// exclude=alpha, q=b so beta matches but alpha is excluded
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/tag-suggestions?exclude=alpha&q=b", nil)
	w := httptest.NewRecorder()
	app.handleTagSuggestions(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.NotContains(t, body, "alpha")
	assert.Contains(t, body, `data-value="beta"`)
}

func TestHandleTagSuggestions_EmptyQuery(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	tags := []types.Tag{
		{Name: "alpha", CreatedAt: now, UpdatedAt: now},
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

	body := w.Body.String()
	require.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, body)
}

func TestHandleTagSuggestions_SortedByPostCount(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	pc1, pc2, pc3 := 5, 100, 50
	tags := []types.Tag{
		{Name: "rare", PostCount: &pc1, CreatedAt: now, UpdatedAt: now},
		{Name: "popular", PostCount: &pc2, CreatedAt: now, UpdatedAt: now},
		{Name: "medium", PostCount: &pc3, CreatedAt: now, UpdatedAt: now},
	}

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/tags") {
			jsonResponse(w, http.StatusOK, client.TagsResponse{Items: &tags})
			return
		}
		http.NotFound(w, r)
	}))

	// "popular" and "rare" both contain "a", "medium" does not
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/tag-suggestions?q=a", nil)
	w := httptest.NewRecorder()
	app.handleTagSuggestions(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	popIdx := strings.Index(body, "popular")
	rareIdx := strings.Index(body, "rare")
	require.GreaterOrEqual(t, popIdx, 0, "expected popular in response: %q", body)
	require.GreaterOrEqual(t, rareIdx, 0, "expected rare in response: %q", body)
	assert.Less(t, popIdx, rareIdx, "expected popular (100 posts) before rare (5 posts)")
}

func TestHandleTagSuggestions_CappedAt20(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()

	tags := make([]types.Tag, 30)
	for i := range tags {
		pc := 30 - i
		tags[i] = types.Tag{Name: fmt.Sprintf("atag-%02d", i), PostCount: &pc, CreatedAt: now, UpdatedAt: now}
	}

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/tags") {
			jsonResponse(w, http.StatusOK, client.TagsResponse{Items: &tags})
			return
		}
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/tag-suggestions?q=atag", nil)
	w := httptest.NewRecorder()
	app.handleTagSuggestions(w, req)

	body := w.Body.String()
	resultCount := strings.Count(body, "ac-item")
	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 20, resultCount)
}
