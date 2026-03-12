package api

import (
	"github.com/dharmab/hyperboard/internal/db/store"
	"github.com/dharmab/hyperboard/internal/storage"
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
