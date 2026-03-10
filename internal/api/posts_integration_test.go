package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dharmab/hyperboard/internal/db/models"
	pkgtypes "github.com/dharmab/hyperboard/pkg/types"
	"github.com/gofrs/uuid/v5"
)

func insertTestPost(t *testing.T, opts ...func(*models.PostSetter)) *models.Post {
	t.Helper()
	ctx := t.Context()
	id := uuid.Must(uuid.NewV4())
	mime := "image/webp"
	contentURL := "http://fake-storage/posts/" + id.String() + "/content.webp"
	thumbnailURL := "http://fake-storage/posts/" + id.String() + "/thumbnail.webp"
	sha := id.String() // unique per test
	now := new(time.Now().UTC())
	hasAudio := false
	setter := &models.PostSetter{
		ID:           &id,
		MimeType:     &mime,
		ContentURL:   &contentURL,
		ThumbnailURL: &thumbnailURL,
		HasAudio:     &hasAudio,
		Sha256:       &sha,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	for _, opt := range opts {
		opt(setter)
	}
	post, err := models.Posts.Insert(setter).One(ctx, testDB)
	if err != nil {
		t.Fatalf("failed to insert test post: %v", err)
	}
	return post
}

func TestGetPost(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	post := insertTestPost(t)

	t.Run("existing post", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/posts/"+post.ID.String(), nil)
		w := httptest.NewRecorder()
		srv.GetPost(w, req, pkgtypes.ID(post.ID))

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
		}

		var got pkgtypes.Post
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if uuid.UUID(got.ID) != post.ID {
			t.Errorf("ID = %v, want %v", got.ID, post.ID)
		}
	})

	t.Run("nonexistent post", func(t *testing.T) {
		fakeID := pkgtypes.ID(uuid.Must(uuid.NewV4()))
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/posts/"+uuid.UUID(fakeID).String(), nil)
		w := httptest.NewRecorder()
		srv.GetPost(w, req, fakeID)

		if w.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestPutPost(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	post := insertTestPost(t)
	postID := pkgtypes.ID(post.ID)

	tagName := "put-test-tag-" + uuid.Must(uuid.NewV4()).String()[:8]

	body := pkgtypes.Post{
		ID:           postID,
		MimeType:     post.MimeType,
		ContentUrl:   post.ContentURL,
		ThumbnailUrl: post.ThumbnailURL,
		Note:         "Updated note",
		Tags:         []pkgtypes.TagName{tagName},
	}
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/posts/"+post.ID.String(), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.PutPost(w, req, postID)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	var got pkgtypes.Post
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if got.Note != "Updated note" {
		t.Errorf("Note = %q, want %q", got.Note, "Updated note")
	}
	if len(got.Tags) != 1 || got.Tags[0] != tagName {
		t.Errorf("Tags = %v, want [%q]", got.Tags, tagName)
	}
}

func TestDeletePost(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	post := insertTestPost(t)
	postID := pkgtypes.ID(post.ID)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/api/v1/posts/"+post.ID.String(), nil)
	w := httptest.NewRecorder()
	srv.DeletePost(w, req, postID)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
	}

	// Verify deleted
	getReq := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/posts/"+post.ID.String(), nil)
	getW := httptest.NewRecorder()
	srv.GetPost(getW, getReq, postID)
	if getW.Code != http.StatusNotFound {
		t.Fatalf("post still found after delete, status = %d", getW.Code)
	}
}

func TestGetPostsSearch(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)

	// Create posts with specific tags
	post1 := insertTestPost(t)
	post2 := insertTestPost(t)

	tag1Name := "search-tag1-" + uuid.Must(uuid.NewV4()).String()[:8]
	tag2Name := "search-tag2-" + uuid.Must(uuid.NewV4()).String()[:8]

	// Tag post1 with tag1
	tagPost(t, srv, post1.ID, tag1Name)
	// Tag post2 with tag2
	tagPost(t, srv, post2.ID, tag2Name)

	t.Run("search by tag inclusion", func(t *testing.T) {
		search := tag1Name
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/posts?search="+search, nil)
		w := httptest.NewRecorder()
		srv.GetPosts(w, req, GetPostsParams{Search: &search})

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
		}

		var resp PostsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if resp.Items == nil {
			t.Fatal("items is nil")
		}
		found := false
		for _, p := range *resp.Items {
			if uuid.UUID(p.ID) == post1.ID {
				found = true
			}
		}
		if !found {
			t.Error("expected post1 in results")
		}
	})

	t.Run("search by tag exclusion", func(t *testing.T) {
		search := "-" + tag1Name
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/posts?search="+search, nil)
		w := httptest.NewRecorder()
		srv.GetPosts(w, req, GetPostsParams{Search: &search})

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
		}

		var resp PostsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if resp.Items != nil {
			for _, p := range *resp.Items {
				if uuid.UUID(p.ID) == post1.ID {
					t.Error("post1 should not appear when excluding its tag")
				}
			}
		}
	})

	t.Run("search untagged", func(t *testing.T) {
		untaggedPost := insertTestPost(t)
		search := "tagged:false"
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/posts?search="+search, nil)
		w := httptest.NewRecorder()
		srv.GetPosts(w, req, GetPostsParams{Search: &search})

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}

		var resp PostsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if resp.Items == nil {
			t.Fatal("items is nil")
		}
		found := false
		for _, p := range *resp.Items {
			if uuid.UUID(p.ID) == untaggedPost.ID {
				found = true
			}
		}
		if !found {
			t.Error("expected untagged post in results")
		}
	})

	t.Run("search type image", func(t *testing.T) {
		search := "type:image"
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/posts?search="+search, nil)
		w := httptest.NewRecorder()
		srv.GetPosts(w, req, GetPostsParams{Search: &search})

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("pagination", func(t *testing.T) {
		limit := 1
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/posts?limit=1", nil)
		w := httptest.NewRecorder()
		srv.GetPosts(w, req, GetPostsParams{Limit: &limit})

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}

		var resp PostsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if resp.Items == nil || len(*resp.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(*resp.Items))
		}
		if resp.Cursor == nil {
			t.Error("expected cursor for next page")
		}
	})
}

func TestGetPostsSortRandom(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)

	// Insert enough posts to require pagination
	for range 3 {
		insertTestPost(t)
	}

	t.Run("first page", func(t *testing.T) {
		search := "sort:random"
		limit := 2
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/posts?search="+search+"&limit=2", nil)
		w := httptest.NewRecorder()
		srv.GetPosts(w, req, GetPostsParams{Search: &search, Limit: &limit})

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
		}

		var resp PostsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if resp.Items == nil || len(*resp.Items) != 2 {
			t.Fatalf("expected 2 items, got %v", resp.Items)
		}
		if resp.Cursor == nil {
			t.Error("expected cursor for next page")
		}

		t.Run("second page", func(t *testing.T) {
			req2 := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/posts?search="+search+"&limit=2&cursor="+*resp.Cursor, nil)
			w2 := httptest.NewRecorder()
			srv.GetPosts(w2, req2, GetPostsParams{Search: &search, Limit: &limit, Cursor: resp.Cursor})

			if w2.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d; body = %s", w2.Code, http.StatusOK, w2.Body.String())
			}

			var resp2 PostsResponse
			if err := json.NewDecoder(w2.Body).Decode(&resp2); err != nil {
				t.Fatalf("failed to decode: %v", err)
			}
			if resp2.Items == nil || len(*resp2.Items) == 0 {
				t.Fatal("expected items on second page")
			}
		})
	})
}

func tagPost(t *testing.T, srv *Server, postID uuid.UUID, tagName string) {
	t.Helper()
	ctx := t.Context()

	// Create the tag
	now := new(time.Now().UTC())
	tag, err := models.Tags.Insert(&models.TagSetter{
		Name:      &tagName,
		CreatedAt: now,
		UpdatedAt: now,
	}).One(ctx, testDB)
	if err != nil {
		t.Fatalf("failed to create tag: %v", err)
	}

	// Get the post model for AttachTags
	post, err := models.FindPost(ctx, testDB, postID)
	if err != nil {
		t.Fatalf("failed to find post: %v", err)
	}
	if err := post.AttachTags(ctx, testDB, tag); err != nil {
		t.Fatalf("failed to attach tag: %v", err)
	}
}
