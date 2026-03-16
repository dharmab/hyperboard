package api

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dharmab/hyperboard/internal/db/store"
	"github.com/dharmab/hyperboard/internal/media"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/gofrs/uuid/v5"
	"github.com/rs/zerolog"
)

// UploadPost handles uploading new media content as a post.
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

	existing, err := s.sqlStore.FindPostBySHA256(ctx, hashHex)
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
	// GIFs are excluded from perceptual hashing and similarity detection.
	var phashVal *sql.Null[int64]
	if contentMIME != media.MIMEGif {
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

	now := time.Now().UTC()
	var phash sql.Null[int64]
	if phashVal != nil {
		phash = *phashVal
	}
	model, err := s.sqlStore.CreatePost(ctx, store.CreatePostInput{
		ID:           postID,
		MimeType:     contentMIME,
		ContentURL:   contentURL,
		ThumbnailURL: thumbnailURL,
		HasAudio:     hasAudioVal,
		SHA256:       hashHex,
		Phash:        phash,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		logger.Error().Err(err).Msg("failed to insert post into database")
		respondWithError(w, http.StatusInternalServerError, "Failed to store post: %v", err)
		return
	}

	logger.Info().Str("mime", contentMIME).Msg("post uploaded")
	model.Tags = nil
	respond(w, http.StatusCreated, postFromModel(model))
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

	now := time.Now().UTC()
	var phash sql.Null[int64]
	if phashVal != nil {
		phash = *phashVal
	}
	model, err := s.sqlStore.UpdatePostContent(ctx, postID, store.UpdatePostContentInput{
		MimeType:     contentMIME,
		ContentURL:   contentURL,
		ThumbnailURL: thumbnailURL,
		HasAudio:     hasAudioVal,
		SHA256:       hashHex,
		Phash:        phash,
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

func (s *Server) RegeneratePostThumbnail(w http.ResponseWriter, r *http.Request, id Id) {
	ctx := r.Context()
	logger := zerolog.Ctx(ctx).With().Stringer("post_id", uuid.UUID(id)).Logger()

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

	contentKey := storageKeyForContent(postID, post.MimeType)
	obj, err := s.mediaStore.Download(ctx, contentKey)
	if err != nil {
		logger.Error().Err(err).Str("key", contentKey).Msg("failed to download content from storage")
		respondWithError(w, http.StatusInternalServerError, "Failed to download content")
		return
	}
	data, err := io.ReadAll(obj.Body)
	_ = obj.Body.Close()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to read content")
		return
	}

	var thumbnailData []byte
	if strings.HasPrefix(post.MimeType, "image/") {
		_, _, thumbnailData, err = media.ProcessImage(data, post.MimeType)
		if err != nil {
			logger.Error().Err(err).Msg("failed to process image for thumbnail regeneration")
			respondWithError(w, http.StatusUnprocessableEntity, "Failed to process image: %v", err)
			return
		}
	} else if strings.HasPrefix(post.MimeType, "video/") {
		thumbnailData, err = media.RegenerateVideoThumbnail(data)
		if err != nil {
			logger.Error().Err(err).Msg("failed to process video for thumbnail regeneration")
			respondWithError(w, http.StatusUnprocessableEntity, "Failed to process video: %v", err)
			return
		}
	} else {
		respondWithError(w, http.StatusUnprocessableEntity, "Cannot regenerate thumbnail for MIME type: %s", post.MimeType)
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

	logger.Info().Msg("thumbnail regenerated")
	respond(w, http.StatusOK, postFromModel(model))
}
