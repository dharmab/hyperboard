package api

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"uuid"

	"github.com/dharmab/hyperboard/internal/db/store"
	"github.com/dharmab/hyperboard/internal/media"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/rs/zerolog"
)

const (
	maxMediaBody      int64 = 4 << 30 // 4 GiB
	postUnlockTimeout       = 10 * time.Second
)

func readMediaRequestBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	if r.ContentLength > maxMediaBody {
		return nil, &http.MaxBytesError{Limit: maxMediaBody}
	}
	return io.ReadAll(http.MaxBytesReader(w, r.Body, maxMediaBody))
}

func respondToMediaBodyReadError(w http.ResponseWriter, err error) {
	if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
		respondWithError(w, http.StatusRequestEntityTooLarge, "Media body exceeds the 4 GiB limit")
		return
	}
	respondWithError(w, http.StatusInternalServerError, "Failed to read request body")
}

func respondToMediaProcessingError(w http.ResponseWriter, err error, message string) {
	status := http.StatusInternalServerError
	if errors.Is(err, media.ErrInvalidMedia) {
		status = http.StatusUnprocessableEntity
	}
	respondWithError(w, status, "%s", message)
}

func validateMediaContentType(r *http.Request, imageOnly bool) (string, error) {
	rawContentType := r.Header.Get("Content-Type")
	if rawContentType == "" {
		return "", errors.New("Content-Type header is required")
	}

	mediaType, _, err := mime.ParseMediaType(rawContentType)
	if err != nil {
		return "", fmt.Errorf("unsupported media type: %s", rawContentType)
	}
	mediaType = strings.ToLower(mediaType)
	if strings.HasPrefix(mediaType, "image/") {
		return mediaType, nil
	}
	if !imageOnly && strings.HasPrefix(mediaType, "video/") {
		return mediaType, nil
	}
	if imageOnly {
		return "", fmt.Errorf("thumbnail must be an image, got: %s", mediaType)
	}
	return "", fmt.Errorf("unsupported media type: %s", mediaType)
}

func (s *Server) acquirePostMutationLock(ctx context.Context, postID uuid.UUID, logger zerolog.Logger) (func(), error) {
	lock, err := s.sqlStore.AcquirePostMutationLock(ctx, postID)
	if err != nil {
		return nil, err
	}
	return func() {
		unlockCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), postUnlockTimeout)
		defer cancel()
		if err := lock.Unlock(unlockCtx); err != nil {
			logger.Error().Err(err).Msg("failed to release post mutation lock")
		}
	}, nil
}

func doesMediaBodyMatchDeclaredMIME(data []byte, declaredMIME string) bool {
	detectedMIME := http.DetectContentType(data)
	if !strings.HasPrefix(detectedMIME, "image/") && !strings.HasPrefix(detectedMIME, "video/") {
		return true
	}
	return detectedMIME == declaredMIME
}

// UploadPost handles uploading new media content as a post.
func (s *Server) UploadPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := *zerolog.Ctx(ctx)

	mimeStr, err := validateMediaContentType(r, false)
	if err != nil {
		respondWithError(w, http.StatusUnsupportedMediaType, "%v", err)
		return
	}

	data, err := readMediaRequestBody(w, r)
	if err != nil {
		logger.Error().Err(err).Msg("failed to read upload request body")
		respondToMediaBodyReadError(w, err)
		return
	}

	logger.Info().Str("mime", mimeStr).Int("size", len(data)).Msg("processing upload")

	if !doesMediaBodyMatchDeclaredMIME(data, mimeStr) {
		respondWithError(w, http.StatusUnprocessableEntity, "Media body does not match declared Content-Type")
		return
	}

	var contentData []byte
	var contentMIME string
	var thumbnailData []byte
	var hasAudioVal bool

	if strings.HasPrefix(mimeStr, "image/") {
		logger.Info().Str("mime", mimeStr).Msg("processing as image")
		contentData, contentMIME, thumbnailData, err = media.ProcessImage(data, mimeStr)
		if err != nil {
			logger.Error().Err(err).Str("mime", mimeStr).Msg("failed to process image")
			respondToMediaProcessingError(w, err, "Failed to process uploaded image")
			return
		}
	} else if strings.HasPrefix(mimeStr, "video/") {
		logger.Info().Str("mime", mimeStr).Msg("processing as video")
		contentData = data
		contentMIME = mimeStr
		thumbnailData, hasAudioVal, err = media.ProcessVideo(data)
		if err != nil {
			logger.Error().Err(err).Str("mime", mimeStr).Msg("failed to process video")
			respondToMediaProcessingError(w, err, "Failed to process uploaded video")
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
		}
	}

	postID := uuid.NewV4()
	logger = logger.With().Stringer("post_id", postID).Logger()

	ext := mimeToExt(contentMIME)
	contentKey := fmt.Sprintf("posts/%s/content.%s", postID, ext)
	thumbnailKey := fmt.Sprintf("posts/%s/thumbnail.webp", postID)

	contentURL, err := s.mediaStore.Upload(ctx, contentKey, contentData, contentMIME)
	if err != nil {
		logger.Error().Err(err).Str("key", contentKey).Msg("failed to upload content to storage")
		respondWithError(w, http.StatusInternalServerError, "Failed to upload post media")
		return
	}

	thumbnailURL, err := s.mediaStore.Upload(ctx, thumbnailKey, thumbnailData, "image/webp")
	if err != nil {
		logger.Error().Err(err).Str("key", thumbnailKey).Msg("failed to upload thumbnail to storage")
		respondWithError(w, http.StatusInternalServerError, "Failed to upload post media")
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
		FileSize:     int64(len(contentData)),
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		logger.Error().Err(err).Msg("failed to insert post into database")
		respondWithError(w, http.StatusInternalServerError, "Failed to create post")
		return
	}

	logger.Info().Str("mime", contentMIME).Msg("post uploaded")
	model.Tags = nil

	var similarItems []types.Post
	if phashVal != nil {
		similar, err := s.sqlStore.FindSimilarPosts(ctx, postID, phashVal.V, 5)
		if err != nil {
			logger.Error().Err(err).Msg("failed to check for similar posts")
		} else if len(similar) > 0 {
			logger.Info().Int("count", len(similar)).Msg("similar posts found")
			similarItems = make([]types.Post, 0, len(similar))
			for _, p := range similar {
				similarItems = append(similarItems, postFromModel(p))
			}
		}
	}

	result := CreatedPostResponse{Post: postFromModel(model)}
	if len(similarItems) > 0 {
		result.Similar = &similarItems
	}
	respond(w, http.StatusCreated, result)
}

func (s *Server) ReplacePostContent(w http.ResponseWriter, r *http.Request, id Id) {
	ctx := r.Context()

	postID := uuid.UUID(id)
	logger := zerolog.Ctx(ctx).With().Stringer("post_id", postID).Logger()

	mimeStr, err := validateMediaContentType(r, false)
	if err != nil {
		respondWithError(w, http.StatusUnsupportedMediaType, "%v", err)
		return
	}

	releaseLock, err := s.acquirePostMutationLock(ctx, postID, logger)
	if err != nil {
		logger.Error().Err(err).Msg("failed to acquire post mutation lock")
		respondWithError(w, http.StatusInternalServerError, "Failed to lock post for media replacement")
		return
	}
	defer releaseLock()

	// Get existing post to determine old storage keys
	existingPost, err := s.sqlStore.GetPost(ctx, postID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to retrieve post for media replacement")
		if errors.Is(err, store.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Post not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve post")
		return
	}

	data, err := readMediaRequestBody(w, r)
	if err != nil {
		respondToMediaBodyReadError(w, err)
		return
	}

	if !doesMediaBodyMatchDeclaredMIME(data, mimeStr) {
		respondWithError(w, http.StatusUnprocessableEntity, "Media body does not match declared Content-Type")
		return
	}

	var contentData []byte
	var contentMIME string
	var thumbnailData []byte
	var hasAudioVal bool

	if strings.HasPrefix(mimeStr, "image/") {
		contentData, contentMIME, thumbnailData, err = media.ProcessImage(data, mimeStr)
		if err != nil {
			logger.Error().Err(err).Str("mime", mimeStr).Msg("failed to process replacement image")
			respondToMediaProcessingError(w, err, "Failed to process replacement image")
			return
		}
	} else if strings.HasPrefix(mimeStr, "video/") {
		contentData = data
		contentMIME = mimeStr
		thumbnailData, hasAudioVal, err = media.ProcessVideo(data)
		if err != nil {
			logger.Error().Err(err).Str("mime", mimeStr).Msg("failed to process replacement video")
			respondToMediaProcessingError(w, err, "Failed to process replacement video")
			return
		}
	} else {
		respondWithError(w, http.StatusUnsupportedMediaType, "Unsupported media type: %s", mimeStr)
		return
	}

	oldContentKey := storageKeyForContent(postID, existingPost.MimeType)
	newContentKey := storageKeyForContent(postID, contentMIME)
	thumbnailKey := storageKeyForThumbnail(postID)

	contentURL, err := s.mediaStore.Upload(ctx, newContentKey, contentData, contentMIME)
	if err != nil {
		logger.Error().Err(err).Str("key", newContentKey).Msg("failed to upload replacement content")
		respondWithError(w, http.StatusInternalServerError, "Failed to replace post media")
		return
	}

	thumbnailURL, err := s.mediaStore.Upload(ctx, thumbnailKey, thumbnailData, "image/webp")
	if err != nil {
		logger.Error().Err(err).Str("key", thumbnailKey).Msg("failed to upload replacement thumbnail")
		respondWithError(w, http.StatusInternalServerError, "Failed to replace post media")
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
		FileSize:     int64(len(contentData)),
		UpdatedAt:    now,
	})
	if err != nil {
		logger.Error().Err(err).Msg("failed to update replacement media in database")
		if errors.Is(err, store.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Post not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to replace post media")
		return
	}
	if oldContentKey != newContentKey {
		if err := s.mediaStore.Delete(ctx, oldContentKey); err != nil {
			logger.Error().Err(err).Str("key", oldContentKey).Msg("failed to delete superseded content object")
		}
	}

	respond(w, http.StatusOK, postFromModel(model))
}

func (s *Server) ReplacePostThumbnail(w http.ResponseWriter, r *http.Request, id Id) {
	ctx := r.Context()
	postID := uuid.UUID(id)
	logger := zerolog.Ctx(ctx).With().Stringer("post_id", postID).Logger()

	mimeStr, err := validateMediaContentType(r, true)
	if err != nil {
		respondWithError(w, http.StatusUnsupportedMediaType, "%v", err)
		return
	}

	releaseLock, err := s.acquirePostMutationLock(ctx, postID, logger)
	if err != nil {
		logger.Error().Err(err).Msg("failed to acquire post mutation lock")
		respondWithError(w, http.StatusInternalServerError, "Failed to lock post for thumbnail replacement")
		return
	}
	defer releaseLock()

	_, err = s.sqlStore.GetPost(ctx, postID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to retrieve post for thumbnail replacement")
		if errors.Is(err, store.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Post not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve post")
		return
	}

	data, err := readMediaRequestBody(w, r)
	if err != nil {
		respondToMediaBodyReadError(w, err)
		return
	}

	if !doesMediaBodyMatchDeclaredMIME(data, mimeStr) {
		respondWithError(w, http.StatusUnprocessableEntity, "Media body does not match declared Content-Type")
		return
	}

	_, _, thumbnailData, err := media.ProcessImage(data, mimeStr)
	if err != nil {
		logger.Error().Err(err).Str("mime", mimeStr).Msg("failed to process replacement thumbnail")
		respondToMediaProcessingError(w, err, "Failed to process thumbnail")
		return
	}

	thumbnailKey := storageKeyForThumbnail(postID)

	thumbnailURL, err := s.mediaStore.Upload(ctx, thumbnailKey, thumbnailData, "image/webp")
	if err != nil {
		logger.Error().Err(err).Str("key", thumbnailKey).Msg("failed to upload replacement thumbnail")
		respondWithError(w, http.StatusInternalServerError, "Failed to replace post thumbnail")
		return
	}

	now := time.Now().UTC()
	model, err := s.sqlStore.UpdatePostThumbnail(ctx, postID, thumbnailURL, now)
	if err != nil {
		logger.Error().Err(err).Msg("failed to update post thumbnail in database")
		if errors.Is(err, store.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Post not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to replace post thumbnail")
		return
	}

	respond(w, http.StatusOK, postFromModel(model))
}

func (s *Server) RegeneratePostThumbnail(w http.ResponseWriter, r *http.Request, id Id) {
	ctx := r.Context()
	logger := zerolog.Ctx(ctx).With().Stringer("post_id", uuid.UUID(id)).Logger()

	postID := uuid.UUID(id)

	releaseLock, err := s.acquirePostMutationLock(ctx, postID, logger)
	if err != nil {
		logger.Error().Err(err).Msg("failed to acquire post mutation lock")
		respondWithError(w, http.StatusInternalServerError, "Failed to lock post for thumbnail regeneration")
		return
	}
	defer releaseLock()

	post, err := s.sqlStore.GetPost(ctx, postID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to retrieve post for thumbnail regeneration")
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
		respondWithError(w, http.StatusInternalServerError, "Failed to regenerate post thumbnail")
		return
	}
	data, err := io.ReadAll(obj.Body)
	_ = obj.Body.Close()
	if err != nil {
		logger.Error().Err(err).Str("key", contentKey).Msg("failed to read content for thumbnail regeneration")
		respondWithError(w, http.StatusInternalServerError, "Failed to load post media")
		return
	}

	var thumbnailData []byte
	if strings.HasPrefix(post.MimeType, "image/") {
		_, _, thumbnailData, err = media.ProcessImage(data, post.MimeType)
		if err != nil {
			logger.Error().Err(err).Msg("failed to process image for thumbnail regeneration")
			respondWithError(w, http.StatusInternalServerError, "Failed to regenerate post thumbnail")
			return
		}
	} else if strings.HasPrefix(post.MimeType, "video/") {
		thumbnailData, err = media.RegenerateVideoThumbnail(data)
		if err != nil {
			logger.Error().Err(err).Msg("failed to process video for thumbnail regeneration")
			respondWithError(w, http.StatusInternalServerError, "Failed to regenerate post thumbnail")
			return
		}
	} else {
		respondWithError(w, http.StatusUnprocessableEntity, "Cannot regenerate thumbnail for MIME type: %s", post.MimeType)
		return
	}

	thumbnailKey := storageKeyForThumbnail(postID)

	thumbnailURL, err := s.mediaStore.Upload(ctx, thumbnailKey, thumbnailData, "image/webp")
	if err != nil {
		logger.Error().Err(err).Str("key", thumbnailKey).Msg("failed to upload regenerated thumbnail")
		respondWithError(w, http.StatusInternalServerError, "Failed to regenerate post thumbnail")
		return
	}

	now := time.Now().UTC()
	model, err := s.sqlStore.UpdatePostThumbnail(ctx, postID, thumbnailURL, now)
	if err != nil {
		logger.Error().Err(err).Msg("failed to update regenerated thumbnail in database")
		if errors.Is(err, store.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Post not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to regenerate post thumbnail")
		return
	}

	logger.Info().Msg("thumbnail regenerated")
	respond(w, http.StatusOK, postFromModel(model))
}
