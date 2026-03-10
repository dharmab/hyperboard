package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dharmab/hyperboard/internal/db/models"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/gofrs/uuid/v5"
)

func TestGetSimilarPosts(t *testing.T) {
	t.Parallel()
	srv := newTestServer(t)

	// Insert a post with a known phash
	var phashVal int64 = 0x0123456789ABCDEF
	post := insertTestPost(t, func(s *models.PostSetter) {
		s.Phash = &sql.Null[int64]{V: phashVal, Valid: true}
	})

	// Insert a similar post (Hamming distance = 1)
	similarPhash := phashVal ^ 1 // flip one bit
	insertTestPost(t, func(s *models.PostSetter) {
		s.Phash = &sql.Null[int64]{V: similarPhash, Valid: true}
	})

	// Insert a dissimilar post (Hamming distance > threshold)
	dissimilarPhash := ^phashVal // flip all bits
	insertTestPost(t, func(s *models.PostSetter) {
		s.Phash = &sql.Null[int64]{V: dissimilarPhash, Valid: true}
	})

	t.Run("returns similar posts", func(t *testing.T) {
		postID := types.ID(post.ID)
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/posts/"+post.ID.String()+"/similar", nil)
		w := httptest.NewRecorder()
		srv.GetSimilarPosts(w, req, postID, GetSimilarPostsParams{})

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
		}

		var resp PostsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if resp.Items == nil || len(*resp.Items) == 0 {
			t.Fatal("expected at least one similar post")
		}
	})

	t.Run("post without phash returns empty", func(t *testing.T) {
		noPhashPost := insertTestPost(t)
		postID := types.ID(noPhashPost.ID)
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/posts/"+noPhashPost.ID.String()+"/similar", nil)
		w := httptest.NewRecorder()
		srv.GetSimilarPosts(w, req, postID, GetSimilarPostsParams{})

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}

		var resp PostsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if resp.Items == nil || len(*resp.Items) != 0 {
			t.Errorf("expected empty items, got %d", len(*resp.Items))
		}
	})

	t.Run("nonexistent post returns not found", func(t *testing.T) {
		fakeID := types.ID(uuid.Must(uuid.NewV4()))
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/posts/"+uuid.UUID(fakeID).String()+"/similar", nil)
		w := httptest.NewRecorder()
		srv.GetSimilarPosts(w, req, fakeID, GetSimilarPostsParams{})

		if w.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}
