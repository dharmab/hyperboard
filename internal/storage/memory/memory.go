package memory

import (
	"context"
	"fmt"
	"io"
	"strings"
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

// Download retrieves data from memory by key.
func (s *Storage) Download(_ context.Context, key string) (*storage.Media, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.objects[key]
	if !ok {
		return nil, fmt.Errorf("object not found: %s", key)
	}
	return &storage.Media{
		Body:          io.NopCloser(strings.NewReader(string(entry.data))),
		ContentType:   entry.contentType,
		ContentLength: int64(len(entry.data)),
	}, nil
}

// Delete removes data from memory by key.
func (s *Storage) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.objects, key)
	return nil
}
