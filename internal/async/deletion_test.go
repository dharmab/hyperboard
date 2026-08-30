package async

import (
	"context"
	"errors"
	"testing"

	"github.com/dharmab/hyperboard/internal/storage"
	"github.com/dharmab/hyperboard/internal/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeDeletionRepository struct {
	post     *deletedPost
	purgedID string
}

func (r *fakeDeletionRepository) NextDeletedPost(context.Context) (*deletedPost, error) {
	return r.post, nil
}

func (r *fakeDeletionRepository) PurgeDeletedPost(_ context.Context, id string) error {
	r.purgedID = id
	return nil
}

type failingDeleteStore struct {
	storage.MediaStore
	failAt  int
	deletes int
}

func (s *failingDeleteStore) Delete(ctx context.Context, key string) error {
	s.deletes++
	if s.deletes == s.failAt {
		return errors.New("delete failed")
	}
	return s.MediaStore.Delete(ctx, key)
}

func TestDeletionControllerReconcile(t *testing.T) {
	t.Parallel()

	mediaStore := memory.New()
	contentURL, err := mediaStore.Upload(t.Context(), "posts/post-id/content.webp", []byte("content"), "image/webp")
	require.NoError(t, err)
	thumbnailURL, err := mediaStore.Upload(t.Context(), "posts/post-id/thumbnail.webp", []byte("thumbnail"), "image/webp")
	require.NoError(t, err)
	repository := &fakeDeletionRepository{post: &deletedPost{ID: "post-id", ContentURL: contentURL, ThumbnailURL: thumbnailURL}}
	controller := deletionController{repository: repository, mediaStore: mediaStore}

	didWork, err := controller.Reconcile(t.Context())
	require.NoError(t, err)
	assert.True(t, didWork)
	assert.Equal(t, "post-id", repository.purgedID)
}

func TestDeletionControllerIsIdleWithoutDeletedPosts(t *testing.T) {
	t.Parallel()

	controller := deletionController{repository: &fakeDeletionRepository{}, mediaStore: memory.New()}
	didWork, err := controller.Reconcile(t.Context())
	require.NoError(t, err)
	assert.False(t, didWork)
}

func TestDeletionControllerRetainsPostWhenMediaDeletionFails(t *testing.T) {
	t.Parallel()

	repository := &fakeDeletionRepository{post: &deletedPost{
		ID:           "post-id",
		ContentURL:   "http://storage/bucket/posts/post-id/content.webp",
		ThumbnailURL: "http://storage/bucket/posts/post-id/thumbnail.webp",
	}}
	mediaStore := &failingDeleteStore{MediaStore: memory.New(), failAt: 2}
	controller := deletionController{repository: repository, mediaStore: mediaStore}

	didWork, err := controller.Reconcile(t.Context())
	assert.False(t, didWork)
	require.EqualError(t, err, "delete thumbnail for post post-id: delete failed")
	assert.Empty(t, repository.purgedID)
}
