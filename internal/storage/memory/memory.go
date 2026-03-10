package memory

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/dharmab/hyperboard/internal/storage"
)

type entry struct {
	data        []byte
	contentType string
}

// Storage is an in-memory storage.Storage implementation for testing.
type Storage struct {
	mu      sync.Mutex
	objects map[string]entry
}

// New creates a new in-memory Storage.
func New() *Storage {
	return &Storage{objects: make(map[string]entry)}
}

func (s *Storage) Upload(_ context.Context, key string, data []byte, contentType string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.objects[key] = entry{data: append([]byte(nil), data...), contentType: contentType}
	return "http://fake-storage/" + key, nil
}

func (s *Storage) Download(_ context.Context, key string) (*storage.Object, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.objects[key]
	if !ok {
		return nil, fmt.Errorf("object not found: %s", key)
	}
	return &storage.Object{
		Body:          io.NopCloser(strings.NewReader(string(entry.data))),
		ContentType:   entry.contentType,
		ContentLength: int64(len(entry.data)),
	}, nil
}

func (s *Storage) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.objects, key)
	return nil
}
