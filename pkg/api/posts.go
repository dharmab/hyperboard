package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/dharmab/hyperboard/internal/db/models"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/gofrs/uuid/v5"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

func postFromModel(model *models.Post) (types.Post, error) {
	post := types.Post{
		ID:           types.ID(model.ID),
		MimeType:     model.MimeType,
		ContentRef:   model.ContentURL,
		ThumbnailRef: model.ThumbnailURL,
		CreatedAt:    model.CreatedAt.V,
		UpdatedAt:    model.UpdatedAt.V,
	}

	// Extract tag names from loaded tags
	tagNames := make([]types.TagName, 0, len(model.R.Tags))
	for _, tag := range model.R.Tags {
		tagNames = append(tagNames, tag.Name)
	}
	post.Tags = tagNames

	return post, nil
}

func parseSearch(search string) types.PostSearch {
	postSearch := types.PostSearch{
		Tags: []types.TagName{},
	}

	if search == "" {
		return postSearch
	}

	// Split search string by whitespace
	parts := strings.Fields(search)
	for _, part := range parts {
		postSearch.Tags = append(postSearch.Tags, part)
	}

	return postSearch
}

func (s *Server) GetPosts(w http.ResponseWriter, r *http.Request, params GetPostsParams) {
	ctx := r.Context()

	mods := []bob.Mod[*dialect.SelectQuery]{
		sm.OrderBy(models.PostColumns.CreatedAt).Desc(),
	}

	if params.Search != nil && *params.Search != "" {
		searchParams := parseSearch(*params.Search)
		if len(searchParams.Tags) > 0 {
			for _, tagName := range searchParams.Tags {
				mods = append(mods, sm.Where(psql.Raw(
					`EXISTS (
						SELECT 1 FROM posts_tags pt
						JOIN tags t ON pt.tag_id = t.id
						WHERE pt.post_id = posts.id AND t.name = ?
					)`, tagName,
				)))
			}
		}
	}

	if params.Cursor != nil && *params.Cursor != "" {
		decodedCursor, err := deobfuscateCursor(*params.Cursor)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid cursor")
			return
		}
		mods = append(mods, sm.Where(models.PostColumns.CreatedAt.LT(psql.Arg(decodedCursor))))
	}

	limit := parseLimit(params.Limit)
	mods = append(mods, sm.Limit(int64(limit+1)))

	posts, err := models.Posts.Query(mods...).All(ctx, s.db)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve posts")
		return
	}

	if err := posts.LoadTags(ctx, s.db); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to load tags")
		return
	}

	hasMore, nextCursor := paginate(len(posts), limit, func() string {
		return posts[limit-1].CreatedAt.V.Format("2006-01-02T15:04:05.999999999Z07:00")
	})
	if hasMore {
		posts = posts[:limit]
	}

	items := make([]types.Post, 0, len(posts))
	for _, post := range posts {
		postResp, err := postFromModel(post)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to convert post")
			return
		}
		items = append(items, postResp)
	}

	resp := PostsResponse{
		Items:  &items,
		Cursor: nextCursor,
	}
	respond(w, http.StatusOK, resp)
}

func (s *Server) GetPost(w http.ResponseWriter, r *http.Request, id Id) {
	ctx := r.Context()

	postID := uuid.UUID(id)

	model, err := models.Posts.Query(
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

	if err := model.LoadTags(ctx, s.db); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to load tags")
		return
	}

	postResp, err := postFromModel(model)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to convert post")
		return
	}

	respond(w, http.StatusOK, postResp)
}

func (s *Server) UploadPost(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "Upload functionality not yet implemented")
}

func (s *Server) PutPost(w http.ResponseWriter, r *http.Request, id Id) {
	ctx := r.Context()

	postID := uuid.UUID(id)

	var post types.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if uuid.UUID(post.ID) != postID {
		respondWithError(w, http.StatusBadRequest, "Post ID mismatch: got %q in body but %q in URL", post.ID, postID)
		return
	}

	existingPost, err := models.Posts.Query(
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

	err = existingPost.Update(ctx, s.db, &models.PostSetter{
		MimeType:     &post.MimeType,
		ContentURL:   &post.ContentRef,
		ThumbnailURL: &post.ThumbnailRef,
		UpdatedAt:    now(),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update post")
		return
	}

	_, err = models.PostsTags.Delete(
		dm.Where(models.PostsTagColumns.PostID.EQ(psql.Arg(postID))),
	).Exec(ctx, s.db)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update post tags")
		return
	}

	for _, tagName := range post.Tags {
		tag, err := models.Tags.Query(
			sm.Where(models.Tags.Name().EQ(psql.Arg(tagName))),
		).One(ctx, s.db)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				respondWithError(w, http.StatusBadRequest, "Tag %q not found", tagName)
				return
			}
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve tag")
			return
		}

		err = existingPost.AttachTags(ctx, s.db, tag)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to attach tag")
			return
		}
	}

	if err := existingPost.LoadTags(ctx, s.db); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to load tags")
		return
	}

	postResp, err := postFromModel(existingPost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to convert post")
		return
	}

	respond(w, http.StatusOK, postResp)
}

func (s *Server) DeletePost(w http.ResponseWriter, r *http.Request, id Id) {
	ctx := r.Context()

	postID := uuid.UUID(id)

	_, err := models.Posts.Delete(
		dm.Where(models.PostColumns.ID.EQ(psql.Arg(postID))),
	).Exec(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Post not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to delete post")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
