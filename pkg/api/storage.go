package api

import "context"

// Storage is the interface for object storage operations.
type Storage interface {
	Upload(ctx context.Context, key string, data []byte, contentType string) (url string, err error)
	Delete(ctx context.Context, key string) error
}
