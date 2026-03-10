package api

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/dharmab/hyperboard/internal/db/store"
	"github.com/dharmab/hyperboard/internal/storage"
	"github.com/rs/zerolog/log"
)

type Server struct {
	sqlStore   store.SQLStore
	mediaStore storage.MediaStore
}

var _ ServerInterface = &Server{}

func NewServer(sqlStore store.SQLStore, mediaStore storage.MediaStore) *Server {
	return &Server{
		sqlStore:   sqlStore,
		mediaStore: mediaStore,
	}
}

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
