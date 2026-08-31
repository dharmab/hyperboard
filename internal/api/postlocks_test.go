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

func TestPostMutationLockWaitersLeavePoolCapacityForHolder(t *testing.T) {
	t.Parallel()

	sqlStore, pool := newTestStoreWithPool(t)
	postID := uuid.NewV4()
	first, err := sqlStore.AcquirePostMutationLock(t.Context(), postID)
	require.NoError(t, err)

	waitCtx, cancel := context.WithCancel(t.Context())
	waiterCount := int(pool.Config().MaxConns) + 1
	started := make(chan struct{}, waiterCount)
	results := make(chan error, waiterCount)
	for range waiterCount {
		go func() {
			started <- struct{}{}
			lock, lockErr := sqlStore.AcquirePostMutationLock(waitCtx, postID)
			if lockErr == nil {
				lockErr = lock.Unlock(context.WithoutCancel(waitCtx))
			}
			results <- lockErr
		}()
	}
	for range waiterCount {
		<-started
	}

	time.Sleep(100 * time.Millisecond)
	require.Zero(t, pool.Stat().AcquiredConns(), "advisory-lock waiters must not consume data pool connections")
	pingCtx, stopPing := context.WithTimeout(t.Context(), time.Second)
	require.NoError(t, sqlStore.Ping(pingCtx), "the lock holder must retain database pool capacity")
	stopPing()

	cancel()
	for range waiterCount {
		require.ErrorIs(t, <-results, context.Canceled)
	}
	require.NoError(t, first.Unlock(t.Context()))
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
