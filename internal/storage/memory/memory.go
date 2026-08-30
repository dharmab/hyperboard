package memory

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/dharmab/hyperboard/internal/storage"
)

// entry holds stored data and content type for an in-memory object.
type entry struct {
	data        []byte
	contentType string
}

// Storage is an in-memory storage.MediaStore implementation for testing.
type Storage struct {
	mu      sync.Mutex
	objects map[string]entry
}

// New creates a new in-memory Storage.
func New() *Storage {
	return &Storage{objects: make(map[string]entry)}
}

// Ping always returns nil (no-op connectivity check).
func (s *Storage) Ping(_ context.Context) error {
	return nil
}

// Upload stores data in memory and returns a fake URL.
func (s *Storage) Upload(_ context.Context, key string, data []byte, contentType string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.objects[key] = entry{data: append([]byte(nil), data...), contentType: contentType}
	return "http://fake-storage/" + key, nil
}

// Size returns the size of an object without retrieving its contents.
func (s *Storage) Size(ctx context.Context, key string) (int64, error) {
	metadata, err := s.Metadata(ctx, key)
	if err != nil {
		return 0, err
	}
	return metadata.ContentLength, nil
}

// Metadata describes an object without retrieving its contents.
func (s *Storage) Metadata(_ context.Context, key string) (*storage.Metadata, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.objects[key]
	if !ok {
		return nil, fmt.Errorf("object not found: %s", key)
	}
	return &storage.Metadata{ContentType: entry.contentType, ContentLength: int64(len(entry.data))}, nil
}

// Download retrieves data from memory by key.
func (s *Storage) Download(_ context.Context, key string) (*storage.Media, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.objects[key]
	if !ok {
		return nil, fmt.Errorf("object not found: %s", key)
	}
	return &storage.Media{
		Body:          io.NopCloser(bytes.NewReader(entry.data)),
		ContentType:   entry.contentType,
		ContentLength: int64(len(entry.data)),
	}, nil
}

// DownloadRange retrieves an inclusive byte range from an object.
func (s *Storage) DownloadRange(_ context.Context, key string, start, end int64) (*storage.Media, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.objects[key]
	if !ok {
		return nil, fmt.Errorf("object not found: %s", key)
	}
	if start < 0 || end < start || end >= int64(len(entry.data)) {
		return nil, fmt.Errorf("invalid byte range %d-%d for object %s", start, end, key)
	}
	return &storage.Media{
		Body:          io.NopCloser(bytes.NewReader(entry.data[start : end+1])),
		ContentType:   entry.contentType,
		ContentLength: end - start + 1,
	}, nil
}

// Delete removes data from memory by key.
func (s *Storage) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.objects, key)
	return nil
}
