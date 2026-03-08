package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
)

type Server struct {
	db      bob.DB
	storage Storage
}

var _ ServerInterface = &Server{}

func NewServer(ctx context.Context, dsn string, storage Storage) (*Server, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	return &Server{
		db:      bob.NewDB(stdlib.OpenDBFromPool(pool)),
		storage: storage,
	}, nil
}

func (s *Server) Close() error {
	return s.db.Close()
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

	obj, err := s.storage.Download(r.Context(), key)
	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("failed to download media")
		http.Error(w, "Media not found", http.StatusNotFound)
		return
	}
	defer obj.Body.Close()

	w.Header().Set("Content-Type", obj.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	if obj.ContentLength > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", obj.ContentLength))
	}
	io.Copy(w, obj.Body)
}
