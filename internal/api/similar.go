package api

import (
	"errors"
	"net/http"

	"github.com/dharmab/hyperboard/internal/db/store"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/gofrs/uuid/v5"
)

// SimilarPostsResponse is returned when an upload is blocked due to
// visually similar posts. It contains both an error message and the
// list of similar posts so the client can display them.
type SimilarPostsResponse struct {
	Message string       `json:"message"`
	Similar []types.Post `json:"similar"`
}

func (s *Server) GetSimilarPosts(w http.ResponseWriter, r *http.Request, id Id, params GetSimilarPostsParams) {
	ctx := r.Context()

	postID := uuid.UUID(id)

	post, err := s.sqlStore.GetPost(ctx, postID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Post not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve post")
		return
	}

	if !post.Phash.Valid {
		respond(w, http.StatusOK, PostsResponse{Items: &[]types.Post{}})
		return
	}

	limit := parseLimit(params.Limit)

	similar, err := s.sqlStore.FindSimilarPosts(ctx, postID, post.Phash.V, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to find similar posts")
		return
	}

	items := make([]types.Post, 0, len(similar))
	for _, p := range similar {
		items = append(items, postFromModel(p))
	}
	respond(w, http.StatusOK, PostsResponse{Items: &items})
}
