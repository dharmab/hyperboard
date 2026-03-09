package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
)

type storageEntry struct {
	data        []byte
	contentType string
}

// FakeStorage is an in-memory Storage implementation for testing.
type FakeStorage struct {
	mu      sync.Mutex
	objects map[string]storageEntry
}

// NewFakeStorage creates a new FakeStorage.
func NewFakeStorage() *FakeStorage {
	return &FakeStorage{objects: make(map[string]storageEntry)}
}

func (f *FakeStorage) Upload(_ context.Context, key string, data []byte, contentType string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.objects[key] = storageEntry{data: append([]byte(nil), data...), contentType: contentType}
	return fmt.Sprintf("http://fake-storage/%s", key), nil
}

func (f *FakeStorage) Download(_ context.Context, key string) (*StorageObject, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	entry, ok := f.objects[key]
	if !ok {
		return nil, fmt.Errorf("object not found: %s", key)
	}
	return &StorageObject{
		Body:          io.NopCloser(strings.NewReader(string(entry.data))),
		ContentType:   entry.contentType,
		ContentLength: int64(len(entry.data)),
	}, nil
}

func (f *FakeStorage) Delete(_ context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.objects, key)
	return nil
}
