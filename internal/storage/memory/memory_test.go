package memory

import (
	"io"
	"testing"
)

func TestStorage_UploadDownloadRoundTrip(t *testing.T) {
	t.Parallel()
	s := New()
	ctx := t.Context()

	data := []byte("hello world")
	contentType := "text/plain"
	_, err := s.Upload(ctx, "key1", data, contentType)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	obj, err := s.Download(ctx, "key1")
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

func TestStorage_DownloadNonexistent(t *testing.T) {
	t.Parallel()
	s := New()
	ctx := t.Context()

	_, err := s.Download(ctx, "missing")
	if err == nil {
		t.Error("expected error for nonexistent key")
	}
}

func TestStorage_DeleteThenDownload(t *testing.T) {
	t.Parallel()
	s := New()
	ctx := t.Context()

	_, err := s.Upload(ctx, "key1", []byte("data"), "text/plain")
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if err := s.Delete(ctx, "key1"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = s.Download(ctx, "key1")
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestStorage_UploadOverwrites(t *testing.T) {
	t.Parallel()
	s := New()
	ctx := t.Context()

	if _, err := s.Upload(ctx, "key1", []byte("first"), "text/plain"); err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if _, err := s.Upload(ctx, "key1", []byte("second"), "application/json"); err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	obj, err := s.Download(ctx, "key1")
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
