package storage

import (
	"context"
	"io"
)

// Media holds the data returned from a Download call.
type Media struct {
	Body          io.ReadCloser
	ContentType   string
	ContentLength int64
}

// MediaStore is the interface for object storage operations.
type MediaStore interface {
	Ping(ctx context.Context) error
	Upload(ctx context.Context, key string, data []byte, contentType string) (url string, err error)
	Download(ctx context.Context, key string) (*Media, error)
	Delete(ctx context.Context, key string) error
}
