package api

import (
	"context"
	"testing"
	"time"

	"uuid"

	"github.com/stretchr/testify/require"
)

func TestPostMutationLockSerializesSamePost(t *testing.T) {
	t.Parallel()

	sqlStore := newTestStore(t)
	postID := uuid.NewV4()
	first, err := sqlStore.AcquirePostMutationLock(t.Context(), postID)
	require.NoError(t, err)

	waitCtx, cancel := context.WithCancel(t.Context())
	started := make(chan struct{})
	result := make(chan error, 1)
	go func() {
		close(started)
		lock, lockErr := sqlStore.AcquirePostMutationLock(waitCtx, postID)
		if lockErr == nil {
			lockErr = lock.Unlock(context.WithoutCancel(waitCtx))
		}
		result <- lockErr
	}()
	<-started
	cancel()

	err = <-result
	require.ErrorIs(t, err, context.Canceled)
	require.NoError(t, first.Unlock(t.Context()))

	second, err := sqlStore.AcquirePostMutationLock(t.Context(), postID)
	require.NoError(t, err, "lock should be acquirable after release")
	require.NoError(t, second.Unlock(t.Context()))
}

func TestPostMutationLockDoesNotSerializeDifferentPosts(t *testing.T) {
	t.Parallel()

	sqlStore := newTestStore(t)
	first, err := sqlStore.AcquirePostMutationLock(t.Context(), uuid.NewV4())
	require.NoError(t, err)
	defer func() { require.NoError(t, first.Unlock(context.Background())) }()

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	second, err := sqlStore.AcquirePostMutationLock(ctx, uuid.NewV4())
	require.NoError(t, err, "different posts should not share a mutation lock")
	require.NoError(t, second.Unlock(t.Context()))
}
