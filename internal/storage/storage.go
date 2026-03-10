package storage

import (
	"context"
	"io"
)

// Object holds the data returned from a Download call.
type Object struct {
	Body          io.ReadCloser
	ContentType   string
	ContentLength int64
}

// Storage is the interface for object storage operations.
type Storage interface {
	Ping(ctx context.Context) error
	Upload(ctx context.Context, key string, data []byte, contentType string) (url string, err error)
	Download(ctx context.Context, key string) (*Object, error)
	Delete(ctx context.Context, key string) error
}
