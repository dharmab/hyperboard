package api

import (
	"context"
	"io"
)

// StorageObject holds the data returned from a Download call.
type StorageObject struct {
	Body          io.ReadCloser
	ContentType   string
	ContentLength int64
}

// Storage is the interface for object storage operations.
type Storage interface {
	Upload(ctx context.Context, key string, data []byte, contentType string) (url string, err error)
	Download(ctx context.Context, key string) (*StorageObject, error)
	Delete(ctx context.Context, key string) error
}
