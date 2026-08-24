package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"uuid"

	"github.com/dharmab/hyperboard/internal/db/models"
	"github.com/dharmab/hyperboard/internal/db/store"
	"github.com/rs/zerolog/log"
)

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

func (s *Server) streamMedia(w http.ResponseWriter, r *http.Request, postID uuid.UUID, key, filename string) {
	obj, err := s.mediaStore.Download(r.Context(), key)
	if err != nil {
		log.Error().Err(err).Stringer("post_id", postID).Msg("failed to download post content")
		respondWithError(w, http.StatusNotFound, "Post content not found")
		return
	}
	defer func() { _ = obj.Body.Close() }()

	w.Header().Set("Content-Type", obj.ContentType)
	w.Header().Set("Cache-Control", "private, max-age=86400")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if filename != "" {
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename=%q`, filename))
	}
	if obj.ContentLength > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(obj.ContentLength, 10))
	}
	_, _ = io.Copy(w, obj.Body)
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
