package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"uuid"

	"github.com/dharmab/hyperboard/internal/db/models"
	"github.com/dharmab/hyperboard/internal/db/store"
	"github.com/dharmab/hyperboard/internal/storage"
	"github.com/rs/zerolog/log"
)

// HeadPostContent returns a post content file's headers without downloading its body.
func (s *Server) HeadPostContent(w http.ResponseWriter, r *http.Request, id Id) {
	post, postID, ok := s.getPostByID(w, r, id)
	if !ok {
		return
	}
	s.serveMediaMetadata(w, r, postID, storageKeyForContent(postID, post.MimeType), "")
}

// GetPostContent streams a post's content file.
func (s *Server) GetPostContent(w http.ResponseWriter, r *http.Request, id Id) {
	post, postID, ok := s.getPostByID(w, r, id)
	if !ok {
		return
	}
	s.streamMedia(w, r, postID, storageKeyForContent(postID, post.MimeType), "")
}

// DownloadPostContent downloads a post's content file.
func (s *Server) DownloadPostContent(w http.ResponseWriter, r *http.Request, id Id) {
	post, postID, ok := s.getPostByID(w, r, id)
	if !ok {
		return
	}
	filename := fmt.Sprintf("%s.%s", postID, mimeToExt(post.MimeType))
	s.streamMedia(w, r, postID, storageKeyForContent(postID, post.MimeType), filename)
}

// GetPostThumbnail streams a post's thumbnail image.
func (s *Server) GetPostThumbnail(w http.ResponseWriter, r *http.Request, id Id) {
	_, postID, ok := s.getPostByID(w, r, id)
	if !ok {
		return
	}
	s.streamMedia(w, r, postID, storageKeyForThumbnail(postID), "")
}

// getPostByID fetches a post by ID. It writes an error response and returns ok=false when the post cannot be fetched.
func (s *Server) getPostByID(w http.ResponseWriter, r *http.Request, id Id) (*models.Post, uuid.UUID, bool) {
	postID := uuid.UUID(id)
	post, err := s.sqlStore.GetPost(r.Context(), postID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Post not found")
			return nil, uuid.Nil(), false
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve post")
		return nil, uuid.Nil(), false
	}
	return post, postID, true
}

func (s *Server) serveMediaMetadata(w http.ResponseWriter, r *http.Request, postID uuid.UUID, key, filename string) {
	metadata, err := s.mediaStore.Metadata(r.Context(), key)
	if err != nil {
		s.respondToMediaStorageError(w, postID, err)
		return
	}
	setMediaHeaders(w, metadata.ContentType, metadata.ContentLength, filename)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) streamMedia(w http.ResponseWriter, r *http.Request, postID uuid.UUID, key, filename string) {
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		metadata, err := s.mediaStore.Metadata(r.Context(), key)
		if err != nil {
			s.respondToMediaStorageError(w, postID, err)
			return
		}
		start, end, ok := parseByteRange(rangeHeader, metadata.ContentLength)
		if !ok {
			w.Header().Set("Accept-Ranges", "bytes")
			w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", metadata.ContentLength))
			respondWithError(w, http.StatusRequestedRangeNotSatisfiable, "Invalid or unsatisfiable byte range")
			return
		}

		obj, err := s.mediaStore.DownloadRange(r.Context(), key, start, end)
		if err != nil {
			s.respondToMediaStorageError(w, postID, err)
			return
		}
		defer func() { _ = obj.Body.Close() }()
		contentLength := end - start + 1
		setMediaHeaders(w, metadata.ContentType, contentLength, filename)
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, metadata.ContentLength))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = io.Copy(w, obj.Body)
		return
	}

	obj, err := s.mediaStore.Download(r.Context(), key)
	if err != nil {
		s.respondToMediaStorageError(w, postID, err)
		return
	}
	defer func() { _ = obj.Body.Close() }()
	setMediaHeaders(w, obj.ContentType, obj.ContentLength, filename)
	_, _ = io.Copy(w, obj.Body)
}

func (s *Server) respondToMediaStorageError(w http.ResponseWriter, postID uuid.UUID, err error) {
	if errors.Is(err, storage.ErrNotFound) {
		respondWithError(w, http.StatusNotFound, "Post content not found")
		return
	}
	log.Error().Err(err).Stringer("post_id", postID).Msg("failed to retrieve post content")
	respondWithError(w, http.StatusInternalServerError, "Failed to retrieve post content")
}

func setMediaHeaders(w http.ResponseWriter, contentType string, contentLength int64, filename string) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Cache-Control", "private, max-age=86400")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if filename != "" {
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename=%q`, filename))
	}
}

// parseByteRange parses value, the complete HTTP Range header value, against
// size, the total length of the stored object in bytes. It accepts one closed
// ("bytes=1-3"), open-ended ("bytes=1-"), or suffix ("bytes=-3") range and
// returns inclusive start and end byte offsets. End offsets and suffix lengths
// beyond size are clamped to the object. The boolean is false when size is not
// positive or value is malformed, contains multiple ranges, or cannot select
// any bytes from the object.
func parseByteRange(value string, size int64) (int64, int64, bool) {
	if size <= 0 || !strings.HasPrefix(value, "bytes=") {
		return 0, 0, false
	}
	spec := strings.TrimPrefix(value, "bytes=")
	if spec == "" || strings.Contains(spec, ",") || strings.TrimSpace(spec) != spec {
		return 0, 0, false
	}
	startText, endText, found := strings.Cut(spec, "-")
	if !found || strings.Contains(endText, "-") {
		return 0, 0, false
	}
	if startText == "" {
		suffixLength, err := strconv.ParseInt(endText, 10, 64)
		if err != nil || suffixLength <= 0 {
			return 0, 0, false
		}
		if suffixLength > size {
			suffixLength = size
		}
		return size - suffixLength, size - 1, true
	}

	start, err := strconv.ParseInt(startText, 10, 64)
	if err != nil || start < 0 || start >= size {
		return 0, 0, false
	}
	if endText == "" {
		return start, size - 1, true
	}
	end, err := strconv.ParseInt(endText, 10, 64)
	if err != nil || end < start {
		return 0, 0, false
	}
	if end >= size {
		end = size - 1
	}
	return start, end, true
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
