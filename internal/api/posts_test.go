package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"sort"
	"testing"
	"time"
	"uuid"

	"github.com/dharmab/hyperboard/internal/db/models"
	"github.com/dharmab/hyperboard/internal/db/store"
	"github.com/dharmab/hyperboard/internal/search"
	"github.com/dharmab/hyperboard/internal/storage/memory"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSearch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  string
		expect search.Query
	}{
		{
			name:  "empty string",
			input: "",
			expect: search.Query{
				IncludedTags: []types.TagName{},
			},
		},
		{
			name:  "single tag",
			input: "landscape",
			expect: search.Query{
				IncludedTags: []types.TagName{"landscape"},
			},
		},
		{
			name:  "multiple tags",
			input: "landscape,portrait",
			expect: search.Query{
				IncludedTags: []types.TagName{"landscape", "portrait"},
			},
		},
		{
			name:  "tags with whitespace",
			input: " landscape , portrait ",
			expect: search.Query{
				IncludedTags: []types.TagName{"landscape", "portrait"},
			},
		},
		{
			name:  "sort created",
			input: "sort:created",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				Sort:         search.SortCreatedAt,
			},
		},
		{
			name:  "sort updated",
			input: "sort:updated",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				Sort:         search.SortUpdatedAt,
			},
		},
		{
			name:  "sort file size",
			input: "sort:file-size",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				Sort:         search.SortFileSize,
			},
		},
		{
			name:  "sort random",
			input: "sort:random",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				Sort:         search.SortRandom,
			},
		},
		{
			name:  "invalid sort ignored",
			input: "sort:invalid",
			expect: search.Query{
				IncludedTags: []types.TagName{},
			},
		},
		{
			name:  "tagged true",
			input: "tagged:true",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				Tagged:       new(true),
			},
		},
		{
			name:  "tagged false",
			input: "tagged:false",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				Tagged:       new(false),
			},
		},
		{
			name:  "tagged true then false uses last value",
			input: "tagged:true,tagged:false",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				Tagged:       new(false),
			},
		},
		{
			name:  "type image",
			input: "type:image",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				TypeImage:    true,
			},
		},
		{
			name:  "type video",
			input: "type:video",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				TypeVideo:    true,
			},
		},
		{
			name:  "type audio",
			input: "type:audio",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				TypeAudio:    true,
			},
		},
		{
			name:  "exclude image type",
			input: "-type:image",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				ExcludeImage: true,
			},
		},
		{
			name:  "exclude video type",
			input: "-type:video",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				ExcludeVideo: true,
			},
		},
		{
			name:  "exclude audio type",
			input: "-type:audio",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				ExcludeAudio: true,
			},
		},
		{
			name:  "order asc",
			input: "order:asc",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				Order:        search.OrderAsc,
			},
		},
		{
			name:  "order desc",
			input: "order:desc",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				Order:        search.OrderDesc,
			},
		},
		{
			name:  "order invalid ignored",
			input: "order:invalid",
			expect: search.Query{
				IncludedTags: []types.TagName{},
			},
		},
		{
			name:  "created_after",
			input: "created_after:2025-01-01T00:00:00Z",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				CreatedAfter: func() *time.Time { t := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC); return &t }(),
			},
		},
		{
			name:  "created_before",
			input: "created_before:2025-06-15T12:30:00Z",
			expect: search.Query{
				IncludedTags:  []types.TagName{},
				CreatedBefore: func() *time.Time { t := time.Date(2025, 6, 15, 12, 30, 0, 0, time.UTC); return &t }(),
			},
		},
		{
			name:  "combined order and created_after with sort",
			input: "sort:created,order:asc,created_after:2025-01-01T00:00:00Z",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				Sort:         search.SortCreatedAt,
				Order:        search.OrderAsc,
				CreatedAfter: func() *time.Time { t := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC); return &t }(),
			},
		},
		{
			name:  "excluded tag",
			input: "-nsfw",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				ExcludedTags: []string{"nsfw"},
			},
		},
		{
			name:  "mixed input",
			input: "landscape,-nsfw,sort:random,tagged:true,type:image",
			expect: search.Query{
				IncludedTags: []types.TagName{"landscape"},
				ExcludedTags: []string{"nsfw"},
				Sort:         search.SortRandom,
				Tagged:       new(true),
				TypeImage:    true,
			},
		},
		{
			name:  "empty terms ignored",
			input: "landscape,,portrait,",
			expect: search.Query{
				IncludedTags: []types.TagName{"landscape", "portrait"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			search := parseSearch(tt.input)
			assert.Equal(t, tt.expect, search)
		})
	}
}

func TestMimeToExt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		mime string
		ext  string
	}{
		{"image/webp", "webp"},
		{"image/jpeg", "jpg"},
		{"image/png", "png"},
		{"image/gif", "gif"},
		{"video/mp4", "mp4"},
		{"video/webm", "webm"},
		{"video/quicktime", "mov"},
		{"application/octet-stream", "bin"},
	}
	for _, tt := range tests {
		t.Run(tt.mime, func(t *testing.T) {
			t.Parallel()
			extension := mimeToExt(tt.mime)
			assert.Equal(t, tt.ext, extension)
		})
	}
}

func TestEncodeDecodeRandomCursor(t *testing.T) {
	t.Parallel()
	original := randomCursor{Seed: 12345, Offset: 64}
	encoded := encodeRandomCursor(original)
	require.NotEmpty(t, encoded)

	var decoded randomCursor
	err := decodeRandomCursor(encoded, &decoded)
	require.NoError(t, err, "decode random cursor")
	assert.Equal(t, original.Seed, decoded.Seed)
	assert.Equal(t, original.Offset, decoded.Offset)
}

func TestDecodeRandomCursorInvalid(t *testing.T) {
	t.Parallel()
	var rc randomCursor
	err := decodeRandomCursor("not-valid-base64!!!", &rc)
	assert.Error(t, err)
}

func FuzzParseSearch(f *testing.F) {
	seeds := []string{
		"",
		"landscape",
		"landscape,portrait",
		" landscape , portrait ",
		"sort:created",
		"sort:updated",
		"sort:file-size",
		"sort:random",
		"sort:invalid",
		"tagged:true",
		"tagged:false",
		"type:image",
		"type:video",
		"type:audio",
		"-type:image",
		"-type:video",
		"-type:audio",
		"-nsfw",
		"landscape,-nsfw,sort:random,tagged:true,type:image",
		"landscape,,portrait,",
		"order:asc",
		"order:desc",
		"order:invalid",
		"created_after:2025-01-01T00:00:00Z",
		"created_before:2025-06-15T12:30:00Z",
		"sort:created,order:asc,created_after:2025-01-01T00:00:00Z",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, input string) {
		parseSearch(input)
	})
}

func createTestPost(t *testing.T, sqlStore store.SQLStore, opts ...func(*store.CreatePostInput)) *models.Post {
	t.Helper()
	ctx := t.Context()
	id := uuid.NewV4()
	mime := "image/webp"
	contentURL := "http://fake-storage/posts/" + id.String() + "/content.webp"
	thumbnailURL := "http://fake-storage/posts/" + id.String() + "/thumbnail.webp"
	sha := id.String() // unique per test
	now := time.Now().UTC()
	hasAudio := false
	input := store.CreatePostInput{
		ID:           id,
		MimeType:     mime,
		ContentURL:   contentURL,
		ThumbnailURL: thumbnailURL,
		HasAudio:     hasAudio,
		SHA256:       sha,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	for _, opt := range opts {
		opt(&input)
	}
	post, err := sqlStore.CreatePost(ctx, input)
	require.NoError(t, err, "insert test post")
	return post
}

func TestGetPost(t *testing.T) {
	t.Parallel()
	sqlStore := newTestStore(t)
	srv := NewServer(sqlStore, memory.New())
	post := createTestPost(t, sqlStore)

	t.Run("existing post", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/posts/"+post.ID.String(), nil)
		w := httptest.NewRecorder()
		srv.GetPost(w, req, types.ID(post.ID))

		responseBody := w.Body.String()
		require.Equalf(t, http.StatusOK, w.Code, "body = %s", responseBody)

		var fetchedPost types.Post
		require.NoError(t, json.NewDecoder(w.Body).Decode(&fetchedPost))
		assert.Equal(t, types.ID(post.ID), fetchedPost.ID)
	})

	t.Run("nonexistent post", func(t *testing.T) {
		t.Parallel()
		fakeID := types.ID(uuid.NewV4())
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/posts/"+uuid.UUID(fakeID).String(), nil)
		w := httptest.NewRecorder()
		srv.GetPost(w, req, fakeID)

		require.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestDownloadPostContent(t *testing.T) {
	t.Parallel()
	sqlStore := newTestStore(t)
	srv := NewServer(sqlStore, memory.New())
	post := createTestPost(t, sqlStore)
	content := []byte("post-content")
	key := storageKeyForContent(uuid.UUID(post.ID), post.MimeType)
	_, err := srv.mediaStore.Upload(t.Context(), key, content, post.MimeType)
	require.NoError(t, err, "failed to upload test content")
	thumbnail := []byte("thumbnail")
	_, err = srv.mediaStore.Upload(t.Context(), storageKeyForThumbnail(uuid.UUID(post.ID)), thumbnail, post.MimeType)
	require.NoError(t, err, "failed to upload test thumbnail")

	t.Run("streams content inline", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/posts/"+post.ID.String()+"/content", nil)
		w := httptest.NewRecorder()
		srv.GetPostContent(w, req, types.ID(post.ID))

		responseBody := w.Body.String()
		require.Equalf(t, http.StatusOK, w.Code, "body = %s", responseBody)
		contentDisposition := w.Header().Get("Content-Disposition")
		assert.Empty(t, contentDisposition)
		responseBytes := w.Body.Bytes()
		assert.Equal(t, content, responseBytes)
	})

	t.Run("downloads content as an attachment", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/posts/"+post.ID.String()+"/content/download", nil)
		w := httptest.NewRecorder()
		srv.DownloadPostContent(w, req, types.ID(post.ID))

		responseBody := w.Body.String()
		require.Equalf(t, http.StatusOK, w.Code, "body = %s", responseBody)
		contentType := w.Header().Get("Content-Type")
		assert.Equal(t, post.MimeType, contentType)
		wantDisposition := `attachment; filename="` + post.ID.String() + `.webp"`
		contentDisposition := w.Header().Get("Content-Disposition")
		assert.Equal(t, wantDisposition, contentDisposition)
		responseBytes := w.Body.Bytes()
		assert.Equal(t, content, responseBytes)
	})

	t.Run("streams thumbnail", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/posts/"+post.ID.String()+"/thumbnail", nil)
		w := httptest.NewRecorder()
		srv.GetPostThumbnail(w, req, types.ID(post.ID))

		responseBody := w.Body.String()
		require.Equalf(t, http.StatusOK, w.Code, "body = %s", responseBody)
		contentType := w.Header().Get("Content-Type")
		assert.Equal(t, post.MimeType, contentType)
		responseBytes := w.Body.Bytes()
		assert.Equal(t, thumbnail, responseBytes)
	})

	t.Run("missing media file", func(t *testing.T) {
		t.Parallel()
		postWithoutMedia := createTestPost(t, sqlStore)
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/posts/"+postWithoutMedia.ID.String()+"/content", nil)
		w := httptest.NewRecorder()
		srv.GetPostContent(w, req, types.ID(postWithoutMedia.ID))

		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("nonexistent post", func(t *testing.T) {
		t.Parallel()
		id := types.ID(uuid.NewV4())
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/posts/"+uuid.UUID(id).String()+"/content", nil)
		w := httptest.NewRecorder()
		srv.GetPostContent(w, req, id)

		require.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestPutPost(t *testing.T) {
	t.Parallel()
	sqlStore := newTestStore(t)
	srv := NewServer(sqlStore, memory.New())
	post := createTestPost(t, sqlStore)
	postID := types.ID(post.ID)

	tagName := "put-test-tag-" + uuid.NewV4().String()[:8]

	body := PostUpdateRequest{
		ID:   postID,
		Note: "Updated note",
		Tags: []types.TagName{tagName},
	}
	b, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/posts/"+post.ID.String(), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.PutPost(w, req, postID)

	responseBody := w.Body.String()
	require.Equalf(t, http.StatusOK, w.Code, "body = %s", responseBody)

	var updatedPost types.Post
	require.NoError(t, json.NewDecoder(w.Body).Decode(&updatedPost))
	assert.Equal(t, "Updated note", updatedPost.Note)
	require.NotNil(t, updatedPost.Tags)
	require.Len(t, updatedPost.Tags, 1)
	assert.Equal(t, tagName, updatedPost.Tags[0])

	t.Run("put post with trailing whitespace tag returns bad request", func(t *testing.T) {
		t.Parallel()
		invalidBody := body
		invalidBody.Tags = []types.TagName{"example "}
		ib, err := json.Marshal(invalidBody)
		require.NoError(t, err)
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/posts/"+post.ID.String(), bytes.NewReader(ib))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.PutPost(w, req, postID)

		responseBody := w.Body.String()
		require.Equalf(t, http.StatusBadRequest, w.Code, "body = %s", responseBody)
	})

	t.Run("put post with zero ID returns bad request", func(t *testing.T) {
		t.Parallel()
		zeroBody := PostUpdateRequest{
			ID:   types.ID(uuid.Nil()),
			Note: "test",
		}
		zb, _ := json.Marshal(zeroBody)
		zReq := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/posts/"+uuid.Nil().String(), bytes.NewReader(zb))
		zReq.Header.Set("Content-Type", "application/json")
		zW := httptest.NewRecorder()
		srv.PutPost(zW, zReq, types.ID(uuid.Nil()))

		responseBody := zW.Body.String()
		require.Equalf(t, http.StatusBadRequest, zW.Code, "body = %s", responseBody)
	})
}

func TestDeletePost(t *testing.T) {
	t.Parallel()
	sqlStore := newTestStore(t)
	srv := NewServer(sqlStore, memory.New())
	post := createTestPost(t, sqlStore)
	postUUID := uuid.UUID(post.ID)
	postID := types.ID(postUUID)
	_, err := srv.mediaStore.Upload(t.Context(), storageKeyForContent(postUUID, post.MimeType), []byte("content"), post.MimeType)
	require.NoError(t, err, "upload content fixture")
	_, err = srv.mediaStore.Upload(t.Context(), storageKeyForThumbnail(postUUID), []byte("thumbnail"), "image/webp")
	require.NoError(t, err, "upload thumbnail fixture")

	req := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/api/v1/posts/"+post.ID.String(), nil)
	w := httptest.NewRecorder()
	srv.DeletePost(w, req, postID)

	require.Equal(t, http.StatusNoContent, w.Code)

	// Verify deleted
	getReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/posts/"+post.ID.String(), nil)
	getW := httptest.NewRecorder()
	srv.GetPost(getW, getReq, postID)
	require.Equal(t, http.StatusNotFound, getW.Code, "post still found after delete")
}

func TestGetPostsSortRandom(t *testing.T) {
	t.Parallel()
	sqlStore := newTestStore(t)
	server := NewServer(sqlStore, memory.New())

	for range 3 {
		createTestPost(t, sqlStore)
	}

	search := "sort:random"
	limit := 2
	firstPageRequest := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/posts?search="+search+"&limit=2", nil)
	firstPageResponse := httptest.NewRecorder()
	server.GetPosts(firstPageResponse, firstPageRequest, GetPostsParams{Search: &search, Limit: &limit})
	firstPageResponseBody := firstPageResponse.Body.String()
	require.Equal(t, http.StatusOK, firstPageResponse.Code, "body = %s", firstPageResponseBody)
	var firstPage PostsResponse
	require.NoError(t, json.NewDecoder(firstPageResponse.Body).Decode(&firstPage))
	require.NotNil(t, firstPage.Items)
	require.Len(t, *firstPage.Items, 2)
	require.NotNil(t, firstPage.Cursor, "expected cursor for next page")

	nextPageRequest := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/posts?search="+search+"&limit=2&cursor="+*firstPage.Cursor, nil)
	nextPageResponse := httptest.NewRecorder()
	server.GetPosts(nextPageResponse, nextPageRequest, GetPostsParams{Search: &search, Limit: &limit, Cursor: firstPage.Cursor})
	nextPageResponseBody := nextPageResponse.Body.String()
	require.Equal(t, http.StatusOK, nextPageResponse.Code, "body = %s", nextPageResponseBody)
	var nextPage PostsResponse
	require.NoError(t, json.NewDecoder(nextPageResponse.Body).Decode(&nextPage))
	require.NotNil(t, nextPage.Items)
	require.Len(t, *nextPage.Items, 1)
	assert.Nil(t, nextPage.Cursor)
}

func TestGetPostsSortByFileSizeWithPagination(t *testing.T) {
	t.Parallel()

	sqlStore := newTestStore(t)
	server := NewServer(sqlStore, memory.New())
	handler := newTestHandlerForServer(t, server)

	want := make([]types.ID, 0, 5)
	for _, fileSize := range []int64{50, 10, 40, 20, 30} {
		post := createTestPost(t, sqlStore, func(input *store.CreatePostInput) {
			input.FileSize = fileSize
		})
		want = append(want, types.ID(post.ID))
	}
	// Ascending file size order for the insertion order above.
	want = []types.ID{want[1], want[3], want[4], want[2], want[0]}

	var got []types.ID
	cursor := ""
	for {
		target := "/api/v1/posts?search=sort:file-size,order:asc&limit=2"
		if cursor != "" {
			target += "&cursor=" + url.QueryEscape(cursor)
		}
		response := performRequest(handler, http.MethodGet, target, "", true)
		responseBody := response.Body.String()
		require.Equal(t, http.StatusOK, response.Code, "body = %s", responseBody)
		var body PostsResponse
		decodeJSON(t, response.Body.Bytes(), &body)
		require.NotNil(t, body.Items)
		for _, post := range *body.Items {
			got = append(got, post.ID)
		}
		if body.Cursor == nil {
			break
		}
		cursor = *body.Cursor
	}

	assert.Equal(t, want, got)
	assertUniquePostIDs(t, got)
}

func TestPostsPagination(t *testing.T) {
	t.Parallel()

	sqlStore := newTestStore(t)
	server := NewServer(sqlStore, memory.New())
	handler := newTestHandlerForServer(t, server)

	baseTime := time.Now().UTC().Add(-time.Hour).Truncate(time.Microsecond)
	want := make([]types.ID, 0, 5)
	for i := range 5 {
		createdAt := baseTime.Add(time.Duration(i) * time.Minute)
		post := createTestPost(t, sqlStore, func(input *store.CreatePostInput) {
			input.CreatedAt = createdAt
			input.UpdatedAt = createdAt
		})

		want = append([]types.ID{types.ID(post.ID)}, want...)
	}

	var postIDs []types.ID
	cursor := ""
	for page := 0; ; page++ {
		target := "/api/v1/posts?limit=2"
		if cursor != "" {
			target += "&cursor=" + url.QueryEscape(cursor)
		}
		response := performRequest(handler, http.MethodGet, target, "", true)
		responseBody := response.Body.String()
		require.Equalf(t, http.StatusOK, response.Code, "page %d status = %d; body = %q", page, response.Code, responseBody)
		var body PostsResponse
		decodeJSON(t, response.Body.Bytes(), &body)
		require.NotNilf(t, body.Items, "page %d items", page)
		require.NotEmptyf(t, *body.Items, "page %d items", page)
		itemCount := len(*body.Items)
		require.LessOrEqualf(t, itemCount, 2, "page %d items", page)
		for _, post := range *body.Items {
			postIDs = append(postIDs, post.ID)
		}
		if body.Cursor == nil {
			break
		}
		cursor = *body.Cursor
	}
	assert.Equalf(t, want, postIDs, "traversed IDs = %v, want %v", postIDs, want)
	assertUniquePostIDs(t, postIDs)

	stale := encodePostCursor(postCursor{Timestamp: baseTime.Add(150 * time.Second).Format(time.RFC3339Nano), ID: uuid.NewV4().String()})
	response := performRequest(handler, http.MethodGet, "/api/v1/posts?limit=2&cursor="+url.QueryEscape(stale), "", true)
	responseBody := response.Body.String()
	require.Equalf(t, http.StatusOK, response.Code, "stale cursor status = %d; body = %q", response.Code, responseBody)
	var body PostsResponse
	decodeJSON(t, response.Body.Bytes(), &body)
	require.NotNil(t, body.Items)
	require.Len(t, *body.Items, 2)
	assert.Equal(t, want[2], (*body.Items)[0].ID)
	assert.Equal(t, want[3], (*body.Items)[1].ID)
}

func assertUniquePostIDs(t *testing.T, ids []types.ID) {
	t.Helper()
	seen := make(map[types.ID]struct{}, len(ids))
	for _, id := range ids {
		assert.NotContains(t, seen, id)
		seen[id] = struct{}{}
	}
}

func TestPostsReadUpdateDelete(t *testing.T) {
	t.Parallel()

	sqlStore := newTestStore(t)
	server := NewServer(sqlStore, memory.New())
	handler := newTestHandlerForServer(t, server)
	model := createTestPost(t, sqlStore)
	postID := types.ID(model.ID)
	postPath := "/api/v1/posts/" + model.ID.String()

	contentKey := storageKeyForContent(uuid.UUID(model.ID), model.MimeType)
	thumbnailKey := storageKeyForThumbnail(uuid.UUID(model.ID))
	_, err := server.mediaStore.Upload(t.Context(), contentKey, []byte("routed post content"), model.MimeType)
	require.NoError(t, err, "upload content fixture")
	_, err = server.mediaStore.Upload(t.Context(), thumbnailKey, []byte("routed post thumbnail"), "image/webp")
	require.NoError(t, err, "upload thumbnail fixture")

	getResponse := performRequest(handler, http.MethodGet, postPath, "", true)
	getResponseBody := getResponse.Body.String()
	require.Equalf(t, http.StatusOK, getResponse.Code, "get status = %d, want %d; body = %q", getResponse.Code, http.StatusOK, getResponseBody)
	getContentType := getResponse.Header().Get("Content-Type")
	assertJSONContentType(t, getContentType)
	var fetched types.Post
	decodeJSON(t, getResponse.Body.Bytes(), &fetched)
	assertPostMatchesModel(t, fetched, model)

	listResponse := performRequest(handler, http.MethodGet, "/api/v1/posts", "", true)
	listResponseBody := listResponse.Body.String()
	require.Equalf(t, http.StatusOK, listResponse.Code, "list status = %d, want %d; body = %q", listResponse.Code, http.StatusOK, listResponseBody)
	listContentType := listResponse.Header().Get("Content-Type")
	assertJSONContentType(t, listContentType)
	var posts PostsResponse
	decodeJSON(t, listResponse.Body.Bytes(), &posts)
	require.NotNil(t, posts.Items)
	require.Len(t, *posts.Items, 1)
	assertPostMatchesModel(t, (*posts.Items)[0], model)

	tagName := "http-post-tag-" + uuid.NewV4().String()[:8]
	note := "Updated through the generated router."
	tags := []types.TagName{tagName}
	updateBody, err := json.Marshal(map[string]any{"id": postID, "note": note, "tags": tags})
	require.NoError(t, err)
	updateResponse := performRequest(handler, http.MethodPut, postPath, string(updateBody), true)
	updateResponseBody := updateResponse.Body.String()
	require.Equalf(t, http.StatusOK, updateResponse.Code, "update status = %d, want %d; body = %q", updateResponse.Code, http.StatusOK, updateResponseBody)
	updateContentType := updateResponse.Header().Get("Content-Type")
	assertJSONContentType(t, updateContentType)
	var updated types.Post
	decodeJSON(t, updateResponse.Body.Bytes(), &updated)
	assert.Equal(t, postID, updated.ID)
	assert.Equal(t, model.MimeType, updated.MimeType)
	assert.Equal(t, model.ContentURL, updated.ContentUrl)
	assert.Equal(t, model.ThumbnailURL, updated.ThumbnailUrl)
	assert.Equal(t, model.HasAudio, updated.HasAudio)
	assert.WithinDuration(t, model.CreatedAt, updated.CreatedAt, 0)
	assert.Equal(t, note, updated.Note)
	assert.Equal(t, tags, updated.Tags)
	updatedAfterFetched := updated.UpdatedAt.After(fetched.UpdatedAt)
	assert.Truef(t, updatedAfterFetched, "updatedAt = %v, want after %v", updated.UpdatedAt, fetched.UpdatedAt)

	persistedResponse := performRequest(handler, http.MethodGet, postPath, "", true)
	persistedResponseBody := persistedResponse.Body.String()
	require.Equalf(t, http.StatusOK, persistedResponse.Code, "get updated status = %d, want %d; body = %q", persistedResponse.Code, http.StatusOK, persistedResponseBody)
	var persisted types.Post
	decodeJSON(t, persistedResponse.Body.Bytes(), &persisted)
	assert.Equal(t, updated, persisted)

	deleteResponse := performRequest(handler, http.MethodDelete, postPath, "", true)
	deleteResponseBody := deleteResponse.Body.String()
	assertEmptyNoContent(t, deleteResponse.Code, deleteResponseBody)

	missingResponse := performRequest(handler, http.MethodGet, postPath, "", true)
	assert.Equalf(t, http.StatusNotFound, missingResponse.Code, "get deleted status = %d, want %d", missingResponse.Code, http.StatusNotFound)
	assertMediaPresent(t, server, contentKey)
	assertMediaPresent(t, server, thumbnailKey)
}

func TestPutPostRejectsBodyPathIDMismatch(t *testing.T) {
	t.Parallel()

	sqlStore := newTestStore(t)
	server := NewServer(sqlStore, memory.New())
	model := createTestPost(t, sqlStore)
	body, err := json.Marshal(map[string]any{"id": types.ID(uuid.NewV4())})
	require.NoError(t, err)
	response := performRequest(
		newTestHandlerForServer(t, server),
		http.MethodPut,
		"/api/v1/posts/"+model.ID.String(),
		string(body),
		true,
	)
	responseBody := response.Body.String()
	require.Equalf(t, http.StatusBadRequest, response.Code, "status = %d, want %d; body = %q", response.Code, http.StatusBadRequest, responseBody)
	contentType := response.Header().Get("Content-Type")
	assertJSONContentType(t, contentType)
	var apiError Error
	decodeJSON(t, response.Body.Bytes(), &apiError)
	assert.NotEmpty(t, apiError.Message)
}

func assertPostMatchesModel(t *testing.T, post types.Post, model *models.Post) {
	t.Helper()
	wantTags := make([]types.TagName, len(model.Tags))
	for i, tag := range model.Tags {
		wantTags[i] = tag.Name
	}
	assert.Equal(t, types.ID(model.ID), post.ID)
	assert.Equal(t, model.MimeType, post.MimeType)
	assert.Equal(t, model.ContentURL, post.ContentUrl)
	assert.Equal(t, model.ThumbnailURL, post.ThumbnailUrl)
	assert.Equal(t, model.Note, post.Note)
	assert.Equal(t, model.HasAudio, post.HasAudio)
	assert.ElementsMatch(t, wantTags, post.Tags)
	assert.WithinDuration(t, model.CreatedAt, post.CreatedAt, 0)
	assert.WithinDuration(t, model.UpdatedAt, post.UpdatedAt, 0)
}

func assertMediaPresent(t *testing.T, server *Server, key string) {
	t.Helper()
	media, err := server.mediaStore.Download(t.Context(), key)
	require.NoError(t, err, "media %q was removed before asynchronous cleanup", key)
	require.NoError(t, media.Body.Close())
}

type searchPostFixture struct {
	id      types.ID
	created time.Time
	updated time.Time
}

func TestPostSearchSortingDatesAndTypes(t *testing.T) {
	t.Parallel()
	sqlStore := newTestStore(t)
	server := NewServer(sqlStore, memory.New())
	handler := newTestHandlerForServer(t, server)
	base := time.Date(2026, time.January, 10, 12, 0, 0, 0, time.UTC)

	fixtures := make([]searchPostFixture, 0, 4)
	add := func(created, updated time.Time, mime string, audio bool) types.ID {
		post := createTestPost(t, sqlStore, func(input *store.CreatePostInput) {
			input.CreatedAt = created
			input.UpdatedAt = updated
			input.MimeType = mime
			input.HasAudio = audio
		})
		id := types.ID(post.ID)
		fixtures = append(fixtures, searchPostFixture{id: id, created: created, updated: updated})
		return id
	}
	imageID := add(base, base.Add(3*time.Hour), "image/webp", false)
	videoID := add(base.Add(time.Hour), base.Add(time.Hour), "video/mp4", true)
	audioImageID := add(base.Add(2*time.Hour), base.Add(2*time.Hour), "image/gif", true)
	tiedID := add(base.Add(3*time.Hour), base.Add(4*time.Hour), "video/webm", false)

	for _, tc := range []struct {
		name   string
		search string
		want   []types.ID
	}{
		{name: "created ascending", search: "sort:created,order:asc", want: sortedFixtureIDs(fixtures, false, true)},
		{name: "created descending", search: "sort:created,order:desc", want: sortedFixtureIDs(fixtures, false, false)},
		{name: "updated ascending", search: "sort:updated,order:asc", want: sortedFixtureIDs(fixtures, true, true)},
		{name: "updated descending", search: "sort:updated,order:desc", want: sortedFixtureIDs(fixtures, true, false)},
		{name: "exclusive date range", search: "sort:created,order:asc,created_after:" + base.Format(time.RFC3339) + ",created_before:" + base.Add(2*time.Hour).Format(time.RFC3339), want: []types.ID{videoID}},
		{name: "images", search: "type:image,sort:created,order:asc", want: []types.ID{imageID, audioImageID}},
		{name: "videos", search: "type:video,sort:created,order:asc", want: []types.ID{videoID, tiedID}},
		{name: "audio", search: "type:audio,sort:created,order:asc", want: []types.ID{videoID, audioImageID}},
		{name: "not images", search: "-type:image,sort:created,order:asc", want: []types.ID{videoID, tiedID}},
		{name: "not videos", search: "-type:video,sort:created,order:asc", want: []types.ID{imageID, audioImageID}},
		{name: "without audio", search: "-type:audio,sort:created,order:asc", want: []types.ID{imageID, tiedID}},
		{name: "image with audio", search: "type:image,type:audio", want: []types.ID{audioImageID}},
		{name: "video without audio", search: "type:video,-type:audio", want: []types.ID{tiedID}},
	} {
		postIDs := searchPostIDs(t, handler, tc.search)
		assert.Equalf(t, tc.want, postIDs, "%s IDs = %v, want %v", tc.name, postIDs, tc.want)
	}
}

func TestPostSearchIncludedExcludedTagsAndAliases(t *testing.T) {
	t.Parallel()
	sqlStore := newTestStore(t)
	server := NewServer(sqlStore, memory.New())
	handler := newTestHandlerForServer(t, server)

	for _, body := range []string{
		`{"name":"animal","aliases":["creature"]}`,
		`{"name":"safe","aliases":["clean"]}`,
		`{"name":"blocked","aliases":["deny"]}`,
	} {
		var input types.Tag
		decodeJSON(t, []byte(body), &input)
		response := performRequest(handler, http.MethodPut, "/api/v1/tags/"+input.Name, body, true)
		responseBody := response.Body.String()
		require.Equalf(t, http.StatusCreated, response.Code, "create tag %q status = %d; body = %q", input.Name, response.Code, responseBody)
	}

	matching := createTestPost(t, sqlStore)
	excluded := createTestPost(t, sqlStore)
	missingIncluded := createTestPost(t, sqlStore)
	for id, tags := range map[uuid.UUID][]string{
		uuid.UUID(matching.ID):        {"animal", "safe"},
		uuid.UUID(excluded.ID):        {"animal", "safe", "blocked"},
		uuid.UUID(missingIncluded.ID): {"animal"},
	} {
		_, err := sqlStore.UpdatePost(t.Context(), id, "", tags, time.Now().UTC())
		require.NoErrorf(t, err, "tag post %s", id)
	}

	postIDs := searchPostIDs(t, handler, "creature,clean,-deny,sort:created")
	expectedPostIDs := []types.ID{types.ID(matching.ID)}
	assert.Equalf(t, expectedPostIDs, postIDs, "combined alias search IDs = %v, want %v", postIDs, expectedPostIDs)
}

func searchPostIDs(t *testing.T, handler http.Handler, search string) []types.ID {
	t.Helper()
	response := performRequest(handler, http.MethodGet, "/api/v1/posts?limit=100&search="+url.QueryEscape(search), "", true)
	responseBody := response.Body.String()
	require.Equalf(t, http.StatusOK, response.Code, "search %q status = %d; body = %q", search, response.Code, responseBody)
	var body PostsResponse
	decodeJSON(t, response.Body.Bytes(), &body)
	require.NotNilf(t, body.Items, "search %q", search)
	ids := make([]types.ID, len(*body.Items))
	for i, post := range *body.Items {
		ids[i] = post.ID
	}
	return ids
}

func sortedFixtureIDs(fixtures []searchPostFixture, updated, ascending bool) []types.ID {
	copyOf := slices.Clone(fixtures)
	sort.Slice(copyOf, func(i, j int) bool {
		left, right := copyOf[i].created, copyOf[j].created
		if updated {
			left, right = copyOf[i].updated, copyOf[j].updated
		}
		if left.Equal(right) {
			if ascending {
				return uuid.UUID(copyOf[i].id).String() < uuid.UUID(copyOf[j].id).String()
			}
			return uuid.UUID(copyOf[i].id).String() > uuid.UUID(copyOf[j].id).String()
		}
		if ascending {
			return left.Before(right)
		}
		return left.After(right)
	})
	ids := make([]types.ID, len(copyOf))
	for i := range copyOf {
		ids[i] = copyOf[i].id
	}
	return ids
}

func TestGetSimilarPosts(t *testing.T) {
	t.Parallel()
	sqlStore := newTestStore(t)
	server := NewServer(sqlStore, memory.New())
	handler := newTestHandlerForServer(t, server)

	// Insert a post with a known PHash.
	var pHash int64 = 0x0123456789ABCDEF
	post := createTestPost(t, sqlStore, func(input *store.CreatePostInput) {
		input.Phash = sql.Null[int64]{V: pHash, Valid: true}
	})

	// Insert a similar post with a PHash at Hamming distance 1.
	similarPost := createTestPost(t, sqlStore, func(input *store.CreatePostInput) {
		input.Phash = sql.Null[int64]{V: pHash ^ 1, Valid: true}
	})

	// Insert a dissimilar post with a PHash beyond the similarity threshold.
	createTestPost(t, sqlStore, func(input *store.CreatePostInput) {
		input.Phash = sql.Null[int64]{V: ^pHash, Valid: true}
	})

	t.Run("returns similar posts", func(t *testing.T) {
		t.Parallel()
		postPath := "/api/v1/posts/" + post.ID.String()
		response := performRequest(handler, http.MethodGet, postPath+"/similar?limit=1", "", true)
		responseBody := response.Body.String()
		require.Equalf(t, http.StatusOK, response.Code, "body = %s", responseBody)
		contentType := response.Header().Get("Content-Type")
		assertJSONContentType(t, contentType)
		var resp PostsResponse
		decodeJSON(t, response.Body.Bytes(), &resp)
		require.NotNil(t, resp.Items)
		require.Len(t, *resp.Items, 1)
		item := (*resp.Items)[0]
		assert.Equal(t, types.ID(similarPost.ID), item.ID)
		assert.NotEmpty(t, item.MimeType)
		assert.NotEmpty(t, item.ContentUrl)
		assert.NotEmpty(t, item.ThumbnailUrl)
		assert.False(t, item.CreatedAt.IsZero())
		assert.False(t, item.UpdatedAt.IsZero())
		assert.Empty(t, item.Tags)
	})

	t.Run("post without PHash returns empty", func(t *testing.T) {
		t.Parallel()
		postPath := "/api/v1/posts/" + createTestPost(t, sqlStore).ID.String()
		response := performRequest(handler, http.MethodGet, postPath+"/similar", "", true)
		require.Equal(t, http.StatusOK, response.Code)
		var resp PostsResponse
		decodeJSON(t, response.Body.Bytes(), &resp)
		require.NotNil(t, resp.Items)
		assert.Empty(t, *resp.Items)
	})

	t.Run("nonexistent post returns not found", func(t *testing.T) {
		t.Parallel()
		fakeID := uuid.NewV4()
		response := performRequest(handler, http.MethodGet, "/api/v1/posts/"+fakeID.String()+"/similar", "", true)
		require.Equal(t, http.StatusNotFound, response.Code)
		contentType := response.Header().Get("Content-Type")
		assertJSONContentType(t, contentType)
	})
}
