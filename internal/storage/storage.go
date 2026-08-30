package storage

import (
	"context"
	"io"
)

// Metadata describes a stored object without retrieving its contents.
type Metadata struct {
	ContentType   string
	ContentLength int64
}

// Media holds the data returned from a download call.
type Media struct {
	Body          io.ReadCloser
	ContentType   string
	ContentLength int64
}

// MediaStore is the interface for object storage operations.
type MediaStore interface {
	Ping(ctx context.Context) error
	Upload(ctx context.Context, key string, data []byte, contentType string) (url string, err error)
	Size(ctx context.Context, key string) (int64, error)
	Metadata(ctx context.Context, key string) (*Metadata, error)
	Download(ctx context.Context, key string) (*Media, error)
	DownloadRange(ctx context.Context, key string, start, end int64) (*Media, error)
	Delete(ctx context.Context, key string) error
}
