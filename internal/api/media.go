package api

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/rs/zerolog/log"
)

// HandleMedia serves media objects from storage.
// Path format: /media/{bucket}/{key...}. The bucket segment is stripped
// since the storage backend already knows its bucket.
func (s *Server) HandleMedia(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/media/")
	// Strip the bucket prefix (first path segment)
	_, key, found := strings.Cut(path, "/")
	if !found || key == "" {
		http.NotFound(w, r)
		return
	}

	obj, err := s.mediaStore.Download(r.Context(), key)
	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("failed to download media")
		http.Error(w, "Media not found", http.StatusNotFound)
		return
	}
	defer func() { _ = obj.Body.Close() }()

	w.Header().Set("Content-Type", obj.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
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
