package api

import (
	"github.com/dharmab/hyperboard/internal/db/store"
	"github.com/dharmab/hyperboard/internal/storage"
)

// Server implements the API server interface with SQL and media store dependencies.
type Server struct {
	sqlStore   store.SQLStore
	mediaStore storage.MediaStore
	// maxOpenRangeBytes caps the length of responses to open-ended range requests.
	maxOpenRangeBytes int64
}

var _ ServerInterface = &Server{}

// NewServer creates a new Server with the given SQL and media stores.
func NewServer(sqlStore store.SQLStore, mediaStore storage.MediaStore) *Server {
	return &Server{
		sqlStore:          sqlStore,
		mediaStore:        mediaStore,
		maxOpenRangeBytes: defaultMaxOpenRangeBytes,
	}
}
