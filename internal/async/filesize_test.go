package async

import (
	"context"
	"errors"
	"testing"

	"github.com/dharmab/hyperboard/internal/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeFileSizeRepository struct {
	post        *fileSizePost
	nextErr     error
	updatedID   string
	updatedSize int64
	updateErr   error
}

func (r *fakeFileSizeRepository) NextPostWithoutFileSize(context.Context) (*fileSizePost, error) {
	return r.post, r.nextErr
}

func (r *fakeFileSizeRepository) SetPostFileSize(_ context.Context, id string, size int64) error {
	r.updatedID = id
	r.updatedSize = size
	return r.updateErr
}

func TestStorageKeyFromURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		mediaURL string
		want     string
	}{
		{
			name:     "existing path-style URL",
			mediaURL: "https://storage.example/bucket/posts/post-id/content.webp",
			want:     "posts/post-id/content.webp",
		},
		{
			name:     "bucket named posts",
			mediaURL: "https://storage.example/posts/posts/post-id/content.webp",
			want:     "posts/post-id/content.webp",
		},
		{
			name:     "endpoint path containing posts",
			mediaURL: "https://storage.example/api/posts/storage/bucket/posts/post-id/content.webp",
			want:     "posts/post-id/content.webp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := storageKeyFromURL(tt.mediaURL)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStorageKeyFromURLRejectsMalformedURLs(t *testing.T) {
	t.Parallel()

	tests := []string{
		"https://storage.example/bucket/not-a-post",
		"https://storage.example/bucket/posts/post-id",
		"https://storage.example/bucket/posts//content.webp",
		"https://storage.example/bucket/posts/post-id/",
		"https://storage.example/bucket/posts/post-id/content.webp/extra",
		"https://storage.example/%zz",
	}

	for _, mediaURL := range tests {
		t.Run(mediaURL, func(t *testing.T) {
			t.Parallel()
			key, err := storageKeyFromURL(mediaURL)
			require.Error(t, err)
			assert.Empty(t, key)
		})
	}
}

func TestFileSizeControllerReconcile(t *testing.T) {
	t.Parallel()

	mediaStore := memory.New()
	content := []byte("post content")
	contentURL, err := mediaStore.Upload(t.Context(), "posts/post-id/content.webp", content, "image/webp")
	require.NoError(t, err)
	repository := &fakeFileSizeRepository{post: &fileSizePost{ID: "post-id", ContentURL: contentURL}}
	controller := fileSizeController{repository: repository, mediaStore: mediaStore}

	didWork, err := controller.Reconcile(t.Context())
	require.NoError(t, err)
	assert.True(t, didWork)
	assert.Equal(t, "post-id", repository.updatedID)
	assert.Equal(t, int64(len(content)), repository.updatedSize)
}

func TestFileSizeControllerIsIdleWhenBackfillIsComplete(t *testing.T) {
	t.Parallel()

	controller := fileSizeController{repository: &fakeFileSizeRepository{}, mediaStore: memory.New()}
	didWork, err := controller.Reconcile(t.Context())
	require.NoError(t, err)
	assert.False(t, didWork)
}

func TestFileSizeControllerRejectsInvalidContentURL(t *testing.T) {
	t.Parallel()

	repository := &fakeFileSizeRepository{post: &fileSizePost{ID: "post-id", ContentURL: "https://storage.example/bucket/not-a-post"}}
	controller := fileSizeController{repository: repository, mediaStore: memory.New()}

	didWork, err := controller.Reconcile(t.Context())
	assert.False(t, didWork)
	require.EqualError(t, err, `determine storage key for post post-id: URL path "/bucket/not-a-post" does not end in posts/<post-id>/<filename>`)
	assert.Empty(t, repository.updatedID)
}

func TestFileSizeControllerReturnsRepositoryError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("database unavailable")
	controller := fileSizeController{
		repository: &fakeFileSizeRepository{nextErr: wantErr},
		mediaStore: memory.New(),
	}

	didWork, err := controller.Reconcile(t.Context())
	assert.False(t, didWork)
	assert.ErrorIs(t, err, wantErr)
}
