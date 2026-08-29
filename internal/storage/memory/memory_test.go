package memory

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorage_UploadDownloadRoundTrip(t *testing.T) {
	t.Parallel()
	s := New()
	ctx := t.Context()

	data := []byte("hello world")
	contentType := "text/plain"
	_, err := s.Upload(ctx, "key1", data, contentType)
	require.NoError(t, err)

	obj, err := s.Download(ctx, "key1")
	require.NoError(t, err)
	defer func() { _ = obj.Body.Close() }()

	got, err := io.ReadAll(obj.Body)
	require.NoError(t, err)
	assert.Equal(t, data, got)
	assert.Equal(t, contentType, obj.ContentType)
	assert.Equal(t, int64(len(data)), obj.ContentLength)
}

func TestStorage_DownloadNonexistent(t *testing.T) {
	t.Parallel()
	s := New()
	ctx := t.Context()

	_, err := s.Download(ctx, "missing")
	assert.Error(t, err)
}

func TestStorage_DeleteThenDownload(t *testing.T) {
	t.Parallel()
	s := New()
	ctx := t.Context()

	_, err := s.Upload(ctx, "key1", []byte("data"), "text/plain")
	require.NoError(t, err)
	deleteErr := s.Delete(ctx, "key1")
	require.NoError(t, deleteErr)

	_, err = s.Download(ctx, "key1")
	assert.Error(t, err)
}

func TestStorage_UploadOverwrites(t *testing.T) {
	t.Parallel()
	s := New()
	ctx := t.Context()

	_, err := s.Upload(ctx, "key1", []byte("first"), "text/plain")
	require.NoError(t, err)
	_, err = s.Upload(ctx, "key1", []byte("second"), "application/json")
	require.NoError(t, err)

	obj, err := s.Download(ctx, "key1")
	require.NoError(t, err)
	defer func() { _ = obj.Body.Close() }()

	got, err := io.ReadAll(obj.Body)
	require.NoError(t, err)
	assert.Equal(t, []byte("second"), got)
	assert.Equal(t, "application/json", obj.ContentType)
}
