package api

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dharmab/hyperboard/internal/db/models"
	"github.com/dharmab/hyperboard/internal/db/store"
	"github.com/dharmab/hyperboard/internal/media"
	"github.com/dharmab/hyperboard/internal/search"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/gofrs/uuid/v5"
	"github.com/rs/zerolog"
)

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

	// Extract tag names from loaded tags
	tagNames := make([]types.TagName, 0, len(model.R.Tags))
	for _, tag := range model.R.Tags {
		tagNames = append(tagNames, tag.Name)
	}
	post.Tags = tagNames

	return post
}

func parseSearch(query string) search.Query {
	postSearch := search.Query{
		IncludedTags: []types.TagName{},
	}

	if query == "" {
		return postSearch
	}

	// Split search string by commas and trim whitespace from each term
	for part := range strings.SplitSeq(query, ",") {
		term := strings.TrimSpace(part)
		if term == "" {
			continue
		}
		if term == string(search.SortRandom) {
			postSearch.Sort = search.SortRandom
		} else if term == string(search.SortCreatedAt) {
			postSearch.Sort = search.SortCreatedAt
		} else if term == string(search.SortUpdatedAt) {
			postSearch.Sort = search.SortUpdatedAt
		} else if strings.HasPrefix(term, "sort:") {
			// Ignore unknown sort values
		} else if term == search.TagTaggedTrue {
			postSearch.Tagged = new(true)
		} else if term == search.TagTaggedFalse {
			postSearch.Tagged = new(false)
		} else if term == search.TagImage {
			postSearch.TypeImage = true
		} else if term == search.TagVideo {
			postSearch.TypeVideo = true
		} else if term == search.TagAudio {
			postSearch.TypeAudio = true
		} else if excluded, ok := strings.CutPrefix(term, "-"); ok && excluded != "" {
			postSearch.ExcludedTags = append(postSearch.ExcludedTags, excluded)
		} else {
			postSearch.IncludedTags = append(postSearch.IncludedTags, term)
		}
	}

	return postSearch
}

type postCursor struct {
	Timestamp string `json:"t"`
	ID        string `json:"id"`
}

func encodePostCursor(pc postCursor) string {
	//nolint:errchkjson // postCursor contains only string fields, json.Marshal cannot fail
	data, _ := json.Marshal(pc)
	return base64.URLEncoding.EncodeToString(data)
}

func decodePostCursor(s string) (postCursor, error) {
	data, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return postCursor{}, err
	}
	var pc postCursor
	return pc, json.Unmarshal(data, &pc)
}

type randomCursor struct {
	Seed   int64 `json:"seed"`
	Offset int   `json:"offset"`
}

func encodeRandomCursor(rc randomCursor) string {
	//nolint:errchkjson // randomCursor contains only primitive fields, json.Marshal cannot fail
	data, _ := json.Marshal(rc)
	return base64.URLEncoding.EncodeToString(data)
}

func decodeRandomCursor(s string, rc *randomCursor) error {
	data, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, rc)
}

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
			items = append(items, postFromModel(post))
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
		items = append(items, postFromModel(post))
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

	respond(w, http.StatusOK, postFromModel(model))
}

func (s *Server) UploadPost(w http.ResponseWriter, r *http.Request, params UploadPostParams) {
	ctx := r.Context()
	logger := *zerolog.Ctx(ctx)

	force := params.Force != nil && *params.Force

	data, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error().Err(err).Msg("failed to read upload request body")
		respondWithError(w, http.StatusInternalServerError, "Failed to read request body: %v", err)
		return
	}

	mimeStr := r.Header.Get("Content-Type")
	if mimeStr == "" {
		respondWithError(w, http.StatusUnsupportedMediaType, "Content-Type header is required")
		return
	}
	// Strip any parameters (e.g. "; charset=utf-8") to get a bare MIME type.
	if idx := strings.Index(mimeStr, ";"); idx != -1 {
		mimeStr = strings.TrimSpace(mimeStr[:idx])
	}

	logger.Info().Str("mime", mimeStr).Int("size", len(data)).Bool("force", force).Msg("processing upload")

	var contentData []byte
	var contentMIME string
	var thumbnailData []byte
	var hasAudioVal bool

	if strings.HasPrefix(mimeStr, "image/") {
		logger.Info().Str("mime", mimeStr).Msg("processing as image")
		contentData, contentMIME, thumbnailData, err = media.ProcessImage(data, mimeStr)
		if err != nil {
			logger.Error().Err(err).Str("mime", mimeStr).Msg("failed to process image")
			respondWithError(w, http.StatusUnprocessableEntity, "Failed to process image: %v", err)
			return
		}
	} else if strings.HasPrefix(mimeStr, "video/") {
		logger.Info().Str("mime", mimeStr).Msg("processing as video")
		contentData = data
		contentMIME = mimeStr
		thumbnailData, hasAudioVal, err = media.ProcessVideo(data)
		if err != nil {
			logger.Error().Err(err).Str("mime", mimeStr).Msg("failed to process video")
			respondWithError(w, http.StatusUnprocessableEntity, "Failed to process video: %v", err)
			return
		}
		logger.Info().Bool("has_audio", hasAudioVal).Msg("video processed")
	} else {
		logger.Info().Str("mime", mimeStr).Msg("unsupported media type")
		respondWithError(w, http.StatusUnsupportedMediaType, "Unsupported media type: %s", mimeStr)
		return
	}

	hash := sha256.Sum256(contentData)
	hashHex := hex.EncodeToString(hash[:])

	existing, err := s.sqlStore.FindPostBySha256(ctx, hashHex)
	if err == nil {
		logger.Info().Stringer("existing_id", existing.ID).Msg("duplicate post detected by sha256")
		respondWithError(w, http.StatusConflict, "Duplicate of existing post %s", existing.ID)
		return
	} else if !errors.Is(err, store.ErrNotFound) {
		logger.Error().Err(err).Msg("failed to check for duplicate post")
		respondWithError(w, http.StatusInternalServerError, "Failed to check for duplicate")
		return
	}

	// Compute perceptual hash from the thumbnail.
	var phashVal *sql.Null[int64]
	pHash, phashErr := media.DhashFromBytes(thumbnailData)
	if phashErr != nil {
		logger.Warn().Err(phashErr).Msg("failed to compute perceptual hash")
	} else {
		phashVal = &sql.Null[int64]{V: pHash, Valid: true}

		// Check for visually similar posts (unless force is set).
		if !force {
			logger.Info().Msg("checking for visually similar posts")
			similar, err := s.sqlStore.FindSimilarPosts(ctx, uuid.Nil, pHash, 5)
			if err != nil {
				logger.Error().Err(err).Msg("failed to check for similar posts")
			} else if len(similar) > 0 {
				logger.Info().Int("count", len(similar)).Msg("similar posts found, rejecting upload")
				items := make([]types.Post, 0, len(similar))
				for _, p := range similar {
					items = append(items, postFromModel(p))
				}
				respond(w, http.StatusConflict, SimilarPostsResponse{
					Message: "Similar posts found",
					Similar: items,
				})
				return
			}
			logger.Info().Msg("no similar posts found")
		} else {
			logger.Info().Msg("skipping similarity check (force=true)")
		}
	}

	postID, err := uuid.NewV4()
	if err != nil {
		logger.Error().Err(err).Msg("failed to generate post ID")
		respondWithError(w, http.StatusInternalServerError, "Failed to generate post ID: %v", err)
		return
	}
	logger = logger.With().Stringer("post_id", postID).Logger()

	ext := mimeToExt(contentMIME)
	contentKey := fmt.Sprintf("posts/%s/content.%s", postID, ext)
	thumbnailKey := fmt.Sprintf("posts/%s/thumbnail.webp", postID)

	contentURL, err := s.mediaStore.Upload(ctx, contentKey, contentData, contentMIME)
	if err != nil {
		logger.Error().Err(err).Str("key", contentKey).Msg("failed to upload content to storage")
		respondWithError(w, http.StatusInternalServerError, "Failed to upload content: %v", err)
		return
	}

	thumbnailURL, err := s.mediaStore.Upload(ctx, thumbnailKey, thumbnailData, "image/webp")
	if err != nil {
		logger.Error().Err(err).Str("key", thumbnailKey).Msg("failed to upload thumbnail to storage")
		respondWithError(w, http.StatusInternalServerError, "Failed to upload thumbnail: %v", err)
		return
	}

	id := postID
	now := new(time.Now().UTC())
	model, err := s.sqlStore.CreatePost(ctx, &models.PostSetter{
		ID:           &id,
		MimeType:     &contentMIME,
		ContentURL:   &contentURL,
		ThumbnailURL: &thumbnailURL,
		HasAudio:     &hasAudioVal,
		Sha256:       &hashHex,
		Phash:        phashVal,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		logger.Error().Err(err).Msg("failed to insert post into database")
		respondWithError(w, http.StatusInternalServerError, "Failed to store post: %v", err)
		return
	}

	logger.Info().Str("mime", contentMIME).Msg("post uploaded")
	model.R.Tags = nil
	respond(w, http.StatusCreated, postFromModel(model))
}

// mimeToExt returns a file extension for a given MIME type.
func mimeToExt(mime string) string {
	switch mime {
	case "image/webp":
		return "webp"
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/gif":
		return "gif"
	case "video/mp4":
		return "mp4"
	case "video/webm":
		return "webm"
	case "video/quicktime":
		return "mov"
	default:
		return "bin"
	}
}

// storageKeyForContent derives the storage key for a post's content from its ID and MIME type.
func storageKeyForContent(postID uuid.UUID, mimeType string) string {
	return fmt.Sprintf("posts/%s/content.%s", postID, mimeToExt(mimeType))
}

// storageKeyForThumbnail derives the storage key for a post's thumbnail from its ID.
func storageKeyForThumbnail(postID uuid.UUID) string {
	return fmt.Sprintf("posts/%s/thumbnail.webp", postID)
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

func (s *Server) ReplacePostContent(w http.ResponseWriter, r *http.Request, id Id) {
	ctx := r.Context()

	postID := uuid.UUID(id)

	// Get existing post to determine old storage keys
	existingPost, err := s.sqlStore.GetPost(ctx, postID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Post not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve post")
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to read request body")
		return
	}

	mimeStr := r.Header.Get("Content-Type")
	if mimeStr == "" {
		respondWithError(w, http.StatusUnsupportedMediaType, "Content-Type header is required")
		return
	}
	if idx := strings.Index(mimeStr, ";"); idx != -1 {
		mimeStr = strings.TrimSpace(mimeStr[:idx])
	}

	var contentData []byte
	var contentMIME string
	var thumbnailData []byte
	var hasAudioVal bool

	if strings.HasPrefix(mimeStr, "image/") {
		contentData, contentMIME, thumbnailData, err = media.ProcessImage(data, mimeStr)
		if err != nil {
			respondWithError(w, http.StatusUnprocessableEntity, "Failed to process image: %v", err)
			return
		}
	} else if strings.HasPrefix(mimeStr, "video/") {
		contentData = data
		contentMIME = mimeStr
		thumbnailData, hasAudioVal, err = media.ProcessVideo(data)
		if err != nil {
			respondWithError(w, http.StatusUnprocessableEntity, "Failed to process video: %v", err)
			return
		}
	} else {
		respondWithError(w, http.StatusUnsupportedMediaType, "Unsupported media type: %s", mimeStr)
		return
	}

	// Delete old content object if the extension changed (new key won't overwrite old one).
	oldContentKey := storageKeyForContent(postID, existingPost.MimeType)
	newContentKey := storageKeyForContent(postID, contentMIME)
	logger := zerolog.Ctx(ctx).With().Stringer("post_id", postID).Logger()
	if oldContentKey != newContentKey {
		logger.Info().Str("old_key", oldContentKey).Str("new_key", newContentKey).Msg("mime type changed, deleting old content object")
		if err := s.mediaStore.Delete(ctx, oldContentKey); err != nil {
			logger.Error().Err(err).Str("key", oldContentKey).Msg("failed to delete old content object")
		}
	}

	thumbnailKey := storageKeyForThumbnail(postID)

	contentURL, err := s.mediaStore.Upload(ctx, newContentKey, contentData, contentMIME)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to upload content")
		return
	}

	thumbnailURL, err := s.mediaStore.Upload(ctx, thumbnailKey, thumbnailData, "image/webp")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to upload thumbnail")
		return
	}

	hashArr := sha256.Sum256(contentData)
	hashHex := hex.EncodeToString(hashArr[:])

	var phashVal *sql.Null[int64]
	pHash, phashErr := media.DhashFromBytes(thumbnailData)
	if phashErr != nil {
		logger.Warn().Err(phashErr).Msg("failed to compute perceptual hash")
	} else {
		phashVal = &sql.Null[int64]{V: pHash, Valid: true}
	}

	now := new(time.Now().UTC())
	model, err := s.sqlStore.UpdatePostContent(ctx, postID, &models.PostSetter{
		MimeType:     &contentMIME,
		ContentURL:   &contentURL,
		ThumbnailURL: &thumbnailURL,
		HasAudio:     &hasAudioVal,
		Sha256:       &hashHex,
		Phash:        phashVal,
		UpdatedAt:    now,
	})
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Post not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to update post")
		return
	}

	respond(w, http.StatusOK, postFromModel(model))
}

func (s *Server) ReplacePostThumbnail(w http.ResponseWriter, r *http.Request, id Id) {
	ctx := r.Context()

	postID := uuid.UUID(id)

	_, err := s.sqlStore.GetPost(ctx, postID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Post not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve post")
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to read request body")
		return
	}

	mimeStr := r.Header.Get("Content-Type")
	if mimeStr == "" {
		respondWithError(w, http.StatusUnsupportedMediaType, "Content-Type header is required")
		return
	}
	if idx := strings.Index(mimeStr, ";"); idx != -1 {
		mimeStr = strings.TrimSpace(mimeStr[:idx])
	}

	if !strings.HasPrefix(mimeStr, "image/") {
		respondWithError(w, http.StatusUnsupportedMediaType, "Thumbnail must be an image, got: %s", mimeStr)
		return
	}

	_, _, thumbnailData, err := media.ProcessImage(data, mimeStr)
	if err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, "Failed to process image: %v", err)
		return
	}

	thumbnailKey := storageKeyForThumbnail(postID)
	thumbnailURL, err := s.mediaStore.Upload(ctx, thumbnailKey, thumbnailData, "image/webp")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to upload thumbnail")
		return
	}

	now := time.Now().UTC()
	model, err := s.sqlStore.UpdatePostThumbnail(ctx, postID, thumbnailURL, now)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Post not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to update post")
		return
	}

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
