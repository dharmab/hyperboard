package storage

import (
	"context"
	"io"
	"testing"
)

func TestFakeStorage_UploadDownloadRoundTrip(t *testing.T) {
	t.Parallel()
	fs := NewFakeStorage()
	ctx := context.Background()

	data := []byte("hello world")
	contentType := "text/plain"
	_, err := fs.Upload(ctx, "key1", data, contentType)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	obj, err := fs.Download(ctx, "key1")
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}
	defer func() { _ = obj.Body.Close() }()

	got, err := io.ReadAll(obj.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("data = %q, want %q", got, data)
	}
	if obj.ContentType != contentType {
		t.Errorf("ContentType = %q, want %q", obj.ContentType, contentType)
	}
	if obj.ContentLength != int64(len(data)) {
		t.Errorf("ContentLength = %d, want %d", obj.ContentLength, len(data))
	}
}

func TestFakeStorage_DownloadNonexistent(t *testing.T) {
	t.Parallel()
	fs := NewFakeStorage()
	ctx := context.Background()

	_, err := fs.Download(ctx, "missing")
	if err == nil {
		t.Error("expected error for nonexistent key")
	}
}

func TestFakeStorage_DeleteThenDownload(t *testing.T) {
	t.Parallel()
	fs := NewFakeStorage()
	ctx := context.Background()

	_, err := fs.Upload(ctx, "key1", []byte("data"), "text/plain")
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if err := fs.Delete(ctx, "key1"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = fs.Download(ctx, "key1")
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestFakeStorage_UploadOverwrites(t *testing.T) {
	t.Parallel()
	fs := NewFakeStorage()
	ctx := context.Background()

	if _, err := fs.Upload(ctx, "key1", []byte("first"), "text/plain"); err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if _, err := fs.Upload(ctx, "key1", []byte("second"), "application/json"); err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	obj, err := fs.Download(ctx, "key1")
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}
	defer func() { _ = obj.Body.Close() }()

	got, _ := io.ReadAll(obj.Body)
	if string(got) != "second" {
		t.Errorf("data = %q, want %q", got, "second")
	}
	if obj.ContentType != "application/json" {
		t.Errorf("ContentType = %q, want %q", obj.ContentType, "application/json")
	}
}
