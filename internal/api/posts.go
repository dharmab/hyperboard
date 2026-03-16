package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dharmab/hyperboard/internal/db/models"
	"github.com/dharmab/hyperboard/internal/db/store"
	"github.com/dharmab/hyperboard/internal/search"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/gofrs/uuid/v5"
	"github.com/rs/zerolog"
)

// postFromModel converts a database Post model to an API Post type.
func postFromModel(model *models.Post) types.Post {
	post := types.Post{
		ID:           types.ID(model.ID),
		MimeType:     model.MimeType,
		ContentUrl:   model.ContentURL,
		ThumbnailUrl: model.ThumbnailURL,
		Note:         model.Note,
		HasAudio:     model.HasAudio,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}

	// Extract tag names and colors from loaded tags
	tagNames := make([]types.TagName, 0, len(model.Tags))
	tagColors := make(map[string]string)
	for _, tag := range model.Tags {
		tagNames = append(tagNames, tag.Name)
		if tag.TagCategory != nil && tag.TagCategory.Color != "" {
			tagColors[tag.Name] = tag.TagCategory.Color
		}
	}
	post.Tags = tagNames
	if len(tagColors) > 0 {
		post.TagColors = &tagColors
	}

	return post
}

// applyCascadingTags sets the CascadingTags field and merges their colors into TagColors.
func applyCascadingTags(post *types.Post, cts []store.CascadingTag) {
	if len(cts) == 0 {
		return
	}
	names := make([]types.TagName, 0, len(cts))
	for _, ct := range cts {
		names = append(names, ct.Name)
		if ct.Color != "" {
			if post.TagColors == nil {
				m := make(map[string]string)
				post.TagColors = &m
			}
			(*post.TagColors)[ct.Name] = ct.Color
		}
	}
	post.CascadingTags = &names
}

// GetPosts handles paginated post listing with search, filtering, and cursor-based pagination.
func (s *Server) GetPosts(w http.ResponseWriter, r *http.Request, params GetPostsParams) {
	ctx := r.Context()
	logger := *zerolog.Ctx(ctx)

	query := ""
	if params.Search != nil {
		query = *params.Search
	}
	searchParams := parseSearch(query)
	logger.Info().
		Str("search", query).
		Strs("tags", searchParams.IncludedTags).
		Strs("exclude_tags", searchParams.ExcludedTags).
		Str("sort", string(searchParams.Sort)).
		Interface("tagged", searchParams.Tagged).
		Bool("type_image", searchParams.TypeImage).
		Bool("type_video", searchParams.TypeVideo).
		Bool("type_audio", searchParams.TypeAudio).
		Msg("parsed search params")

	limit := parseLimit(params.Limit)

	listParams := store.ListPostsParams{
		Query: searchParams,
		Limit: limit,
	}

	if searchParams.Sort == search.SortRandom {
		currentSeed := time.Now().Unix() / 21600
		listParams.RandomSeed = &currentSeed

		if params.Cursor != nil && *params.Cursor != "" {
			var rc randomCursor
			if err := decodeRandomCursor(*params.Cursor, &rc); err == nil {
				if rc.Seed == currentSeed {
					listParams.RandomOffset = rc.Offset
					logger.Info().Int64("seed", currentSeed).Int("offset", rc.Offset).Msg("resuming random cursor")
				} else {
					logger.Info().Int64("old_seed", rc.Seed).Int64("new_seed", currentSeed).Msg("random window rolled, restarting from offset 0")
				}
			}
		}

		posts, hasMore, err := s.sqlStore.ListPosts(ctx, listParams)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve posts")
			return
		}

		var nextCursor *string
		if hasMore {
			rc := randomCursor{Seed: currentSeed, Offset: listParams.RandomOffset + limit}
			encoded := encodeRandomCursor(rc)
			nextCursor = &encoded
		}

		items := make([]types.Post, 0, len(posts))
		for _, post := range posts {
			p := postFromModel(post)
			cts, _ := s.sqlStore.GetPostCascadingTags(ctx, post.ID)
			applyCascadingTags(&p, cts)
			items = append(items, p)
		}
		respond(w, http.StatusOK, PostsResponse{Items: &items, Cursor: nextCursor})
		return
	}

	// Deterministic sort with cursor
	if params.Cursor != nil && *params.Cursor != "" {
		pc, err := decodePostCursor(*params.Cursor)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid cursor")
			return
		}
		ts, err := time.Parse(time.RFC3339Nano, pc.Timestamp)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid cursor")
			return
		}
		cursorID, err := uuid.FromString(pc.ID)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid cursor")
			return
		}
		listParams.CursorTime = &ts
		listParams.CursorID = &cursorID
	}

	posts, hasMore, err := s.sqlStore.ListPosts(ctx, listParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve posts")
		return
	}

	var nextCursor *string
	if hasMore {
		last := posts[len(posts)-1]
		var ts string
		if searchParams.Sort == search.SortUpdatedAt {
			ts = last.UpdatedAt.Format(time.RFC3339Nano)
		} else {
			ts = last.CreatedAt.Format(time.RFC3339Nano)
		}
		encoded := encodePostCursor(postCursor{Timestamp: ts, ID: last.ID.String()})
		nextCursor = &encoded
	}

	items := make([]types.Post, 0, len(posts))
	for _, post := range posts {
		p := postFromModel(post)
		cts, _ := s.sqlStore.GetPostCascadingTags(ctx, post.ID)
		applyCascadingTags(&p, cts)
		items = append(items, p)
	}
	respond(w, http.StatusOK, PostsResponse{Items: &items, Cursor: nextCursor})
}

func (s *Server) GetPost(w http.ResponseWriter, r *http.Request, id Id) {
	ctx := r.Context()

	model, err := s.sqlStore.GetPost(ctx, uuid.UUID(id))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Post not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve post")
		return
	}

	post := postFromModel(model)
	cts, err := s.sqlStore.GetPostCascadingTags(ctx, uuid.UUID(id))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve cascading tags")
		return
	}
	applyCascadingTags(&post, cts)

	respond(w, http.StatusOK, post)
}

func (s *Server) PutPost(w http.ResponseWriter, r *http.Request, id Id) {
	ctx := r.Context()

	postID := uuid.UUID(id)

	var post types.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if uuid.UUID(post.ID) == uuid.Nil {
		respondWithError(w, http.StatusBadRequest, "Post ID is required")
		return
	}

	if uuid.UUID(post.ID) != postID {
		respondWithError(w, http.StatusBadRequest, "Post ID mismatch: got %q in body but %q in URL", post.ID, postID)
		return
	}

	now := time.Now().UTC()
	model, err := s.sqlStore.UpdatePost(ctx, postID, post.Note, post.Tags, now)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Post not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to update post")
		return
	}

	logger := zerolog.Ctx(ctx).With().Stringer("post_id", postID).Logger()
	logger.Info().Int("tag_count", len(post.Tags)).Msg("post updated")
	respond(w, http.StatusOK, postFromModel(model))
}

func (s *Server) DeletePost(w http.ResponseWriter, r *http.Request, id Id) {
	ctx := r.Context()

	postID := uuid.UUID(id)

	// Fetch the post first to get storage keys for cleanup.
	post, err := s.sqlStore.GetPost(ctx, postID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Post not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve post")
		return
	}

	logger := zerolog.Ctx(ctx).With().Stringer("post_id", postID).Logger()
	contentKey := storageKeyForContent(postID, post.MimeType)
	thumbnailKey := storageKeyForThumbnail(postID)
	if err := s.mediaStore.Delete(ctx, contentKey); err != nil {
		logger.Error().Err(err).Str("key", contentKey).Msg("failed to delete content object")
		respondWithError(w, http.StatusInternalServerError, "Failed to delete post content from storage")
		return
	}
	if err := s.mediaStore.Delete(ctx, thumbnailKey); err != nil {
		logger.Error().Err(err).Str("key", thumbnailKey).Msg("failed to delete thumbnail object")
		respondWithError(w, http.StatusInternalServerError, "Failed to delete post thumbnail from storage")
		return
	}

	_, err = s.sqlStore.DeletePost(ctx, postID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete post")
		return
	}

	logger.Info().Msg("post deleted")
	w.WriteHeader(http.StatusNoContent)
}
