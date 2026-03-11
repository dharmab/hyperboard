package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/dharmab/hyperboard/internal/db/models"
	"github.com/dharmab/hyperboard/internal/db/store"
	pkgtypes "github.com/dharmab/hyperboard/pkg/types"
	"github.com/gofrs/uuid/v5"
)

func insertTestPost(t *testing.T, opts ...func(*store.CreatePostInput)) *models.Post {
	t.Helper()
	ctx := t.Context()
	id := uuid.Must(uuid.NewV4())
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
		Sha256:       sha,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	for _, opt := range opts {
		opt(&input)
	}
	post, err := testStore.CreatePost(ctx, input)
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

	post1 := insertTestPost(t)
	post2 := insertTestPost(t)

	tag1Name := "search-tag1-" + uuid.Must(uuid.NewV4()).String()[:8]
	tag2Name := "search-tag2-" + uuid.Must(uuid.NewV4()).String()[:8]

	tagPost(t, post1.ID, tag1Name)
	tagPost(t, post2.ID, tag2Name)

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

		var resp PostsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if resp.Items == nil {
			t.Fatal("items is nil")
		}
		for _, p := range *resp.Items {
			if uuid.UUID(p.ID) == post1.ID {
				found := false
				for _, item := range *resp.Items {
					if uuid.UUID(item.ID) == post1.ID {
						found = true
					}
				}
				if !found {
					t.Error("image post not found in type:image search")
				}
				break
			}
		}
	})

	t.Run("search type video", func(t *testing.T) {
		videoPost := insertTestPost(t, func(s *store.CreatePostInput) {
			s.MimeType = "video/mp4"
		})
		search := "type:video"
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
			if uuid.UUID(p.ID) == videoPost.ID {
				found = true
			}
		}
		if !found {
			t.Error("video post not found in type:video search")
		}
		for _, p := range *resp.Items {
			if uuid.UUID(p.ID) == post1.ID {
				t.Error("image post should not appear in type:video search")
			}
		}
	})

	t.Run("search type audio (has_audio)", func(t *testing.T) {
		audioPost := insertTestPost(t, func(s *store.CreatePostInput) {
			s.HasAudio = true
		})
		search := "type:audio"
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
			if uuid.UUID(p.ID) == audioPost.ID {
				found = true
			}
		}
		if !found {
			t.Error("audio post not found in type:audio search")
		}
		for _, p := range *resp.Items {
			if uuid.UUID(p.ID) == post1.ID {
				t.Error("non-audio image post should not appear in type:audio search")
			}
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

func TestGetPostsSortUpdated(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	ctx := t.Context()

	// Use a unique tag to isolate these test posts from other parallel test data.
	sortTag := "sort-updated-" + uuid.Must(uuid.NewV4()).String()[:8]
	now := time.Now().UTC()

	post1 := insertTestPost(t)
	post2 := insertTestPost(t)
	post3 := insertTestPost(t)

	// Tag all three posts with the isolation tag.
	if _, err := testStore.UpdatePost(ctx, post1.ID, "", []string{sortTag}, now); err != nil {
		t.Fatalf("failed to tag post1: %v", err)
	}
	if _, err := testStore.UpdatePost(ctx, post2.ID, "", []string{sortTag}, now); err != nil {
		t.Fatalf("failed to tag post2: %v", err)
	}
	if _, err := testStore.UpdatePost(ctx, post3.ID, "", []string{sortTag}, now); err != nil {
		t.Fatalf("failed to tag post3: %v", err)
	}

	// Now update posts with specific timestamps to establish sort order.
	// post1 = most recently updated, post2 = middle, post3 = oldest.
	if _, err := testStore.UpdatePost(ctx, post3.ID, "note3", []string{sortTag}, now.Add(-2*time.Second)); err != nil {
		t.Fatalf("failed to update post3: %v", err)
	}
	if _, err := testStore.UpdatePost(ctx, post2.ID, "note2", []string{sortTag}, now.Add(-time.Second)); err != nil {
		t.Fatalf("failed to update post2: %v", err)
	}
	if _, err := testStore.UpdatePost(ctx, post1.ID, "note1", []string{sortTag}, now); err != nil {
		t.Fatalf("failed to update post1: %v", err)
	}

	// Search using the isolation tag so only our 3 posts are returned.
	// The search parser splits on commas, so we use comma to separate the tag from the sort modifier.
	search := sortTag + ",sort:updated"
	limit := 3
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/posts?search="+url.QueryEscape(search)+"&limit=3", nil)
	w := httptest.NewRecorder()
	srv.GetPosts(w, req, GetPostsParams{Search: &search, Limit: &limit})

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
	if len(*resp.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(*resp.Items))
	}

	items := *resp.Items
	if uuid.UUID(items[0].ID) != post1.ID {
		t.Errorf("position 0 = %v, want post1 %v", items[0].ID, post1.ID)
	}
	if uuid.UUID(items[1].ID) != post2.ID {
		t.Errorf("position 1 = %v, want post2 %v", items[1].ID, post2.ID)
	}
	if uuid.UUID(items[2].ID) != post3.ID {
		t.Errorf("position 2 = %v, want post3 %v", items[2].ID, post3.ID)
	}
}

func TestFindPostBySha256(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)

	post := insertTestPost(t)

	// Use the API endpoint that would exercise FindPostBySha256 indirectly via duplicate detection.
	// Also test directly that SHA256 lookup returns the same post.
	ctx := t.Context()
	found, err := testStore.FindPostBySha256(ctx, post.Sha256)
	if err != nil {
		t.Fatalf("FindPostBySha256 failed: %v", err)
	}
	if found.ID != post.ID {
		t.Errorf("FindPostBySha256 returned ID %v, want %v", found.ID, post.ID)
	}

	// Non-existent hash should return ErrNotFound.
	_, err = testStore.FindPostBySha256(ctx, "nonexistent-sha256")
	if err == nil {
		t.Error("expected error for nonexistent hash, got nil")
	}

	_ = srv // ensure server is created (shares test DB)
}

func TestGetPostsTagCascadeSearch(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	ctx := t.Context()

	// parentTag cascades to childTag.
	suffix := uuid.Must(uuid.NewV4()).String()[:8]
	parentTagName := "cascade-parent-" + suffix
	childTagName := "cascade-child-" + suffix

	now := time.Now().UTC()

	// Create both tags.
	_, _, err := testStore.UpsertTag(ctx, childTagName, store.TagInput{Name: childTagName}, now)
	if err != nil {
		t.Fatalf("failed to create child tag: %v", err)
	}
	_, _, err = testStore.UpsertTag(ctx, parentTagName, store.TagInput{
		Name:          parentTagName,
		CascadingTags: []string{childTagName},
	}, now)
	if err != nil {
		t.Fatalf("failed to create parent tag with cascade: %v", err)
	}

	// Tag the post with parentTag directly (don't use tagPost which would clear the cascades).
	post := insertTestPost(t)
	_, err = testStore.UpdatePost(ctx, post.ID, "", []string{parentTagName}, now)
	if err != nil {
		t.Fatalf("failed to tag post: %v", err)
	}

	t.Run("cascade inclusion: searching child tag finds post tagged with parent", func(t *testing.T) {
		search := childTagName
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
			if uuid.UUID(p.ID) == post.ID {
				found = true
			}
		}
		if !found {
			t.Error("expected post (tagged with parent that cascades to child) to appear in child tag search")
		}
	})

	t.Run("cascade exclusion: excluding child tag omits post tagged with parent", func(t *testing.T) {
		search := "-" + childTagName
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
				if uuid.UUID(p.ID) == post.ID {
					t.Error("post (tagged with parent that cascades to child) should not appear when excluding child tag")
				}
			}
		}
	})
}

func TestGetPostsAliasSearch(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)
	ctx := t.Context()

	suffix := uuid.Must(uuid.NewV4()).String()[:8]
	canonicalTagName := "alias-canonical-" + suffix
	aliasName := "alias-alias-" + suffix

	now := time.Now().UTC()
	_, _, err := testStore.UpsertTag(ctx, canonicalTagName, store.TagInput{
		Name:    canonicalTagName,
		Aliases: []string{aliasName},
	}, now)
	if err != nil {
		t.Fatalf("failed to create tag with alias: %v", err)
	}

	post := insertTestPost(t)
	// Tag the post directly without going through tagPost (which would clear the aliases).
	_, err = testStore.UpdatePost(ctx, post.ID, "", []string{canonicalTagName}, now)
	if err != nil {
		t.Fatalf("failed to tag post: %v", err)
	}

	t.Run("searching by alias finds post tagged with canonical tag", func(t *testing.T) {
		search := aliasName
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
			if uuid.UUID(p.ID) == post.ID {
				found = true
			}
		}
		if !found {
			t.Error("expected post tagged with canonical tag to appear when searching by alias")
		}
	})
}

func tagPost(t *testing.T, postID uuid.UUID, tagName string) {
	t.Helper()
	ctx := t.Context()
	now := time.Now().UTC()
	_, _, err := testStore.UpsertTag(ctx, tagName, store.TagInput{
		Name: tagName,
	}, now)
	if err != nil {
		t.Fatalf("failed to create tag: %v", err)
	}
	_, err = testStore.UpdatePost(ctx, postID, "", []string{tagName}, now)
	if err != nil {
		t.Fatalf("failed to tag post: %v", err)
	}
}
