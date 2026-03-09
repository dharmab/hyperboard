package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/dharmab/hyperboard/internal/db/models"
	"github.com/dharmab/hyperboard/internal/types"
	"github.com/gofrs/uuid/v5"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

// SimilarPostsResponse is returned when an upload is blocked due to
// visually similar posts. It contains both an error message and the
// list of similar posts so the client can display them.
type SimilarPostsResponse struct {
	Message string       `json:"message"`
	Similar []types.Post `json:"similar"`
}

// findSimilarPosts returns posts whose perceptual hash is within the
// configured Hamming distance threshold of the given hash.
func (s *Server) findSimilarPosts(ctx context.Context, excludeID uuid.UUID, pHash int64, limit int) (models.PostSlice, error) {
	mods := []bob.Mod[*dialect.SelectQuery]{
		sm.Where(psql.Raw("phash IS NOT NULL")),
		sm.Where(psql.Raw("bit_count((phash # ?)::bit(64)) <= ?", pHash, s.similarityThreshold)),
		sm.OrderBy(psql.Raw("bit_count((phash # ?)::bit(64))", pHash)),
		sm.Limit(int64(limit)),
	}

	if excludeID != uuid.Nil {
		mods = append(mods, sm.Where(models.PostColumns.ID.NE(psql.Arg(excludeID))))
	}

	return models.Posts.Query(mods...).All(ctx, s.db)
}

func (s *Server) GetSimilarPosts(w http.ResponseWriter, r *http.Request, id Id, params GetSimilarPostsParams) {
	ctx := r.Context()

	postID := uuid.UUID(id)

	post, err := models.Posts.Query(
		sm.Where(models.PostColumns.ID.EQ(psql.Arg(postID))),
	).One(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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

	similar, err := s.findSimilarPosts(ctx, postID, post.Phash.V, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to find similar posts")
		return
	}

	if err := similar.LoadTags(ctx, s.db); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to load tags")
		return
	}

	items := make([]types.Post, 0, len(similar))
	for _, p := range similar {
		items = append(items, postFromModel(p))
	}
	respond(w, http.StatusOK, PostsResponse{Items: &items})
}
