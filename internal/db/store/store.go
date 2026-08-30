package store

import (
	"context"
	"errors"

	"uuid"
)

var (
	// ErrNotFound is returned when a requested resource does not exist.
	ErrNotFound = errors.New("not found")
	// ErrConflict is returned when a requested name or alias is already in use.
	ErrConflict = errors.New("name or alias conflict")
)

// SQLStore combines all sub-interfaces for database operations.
type SQLStore interface {
	Pinger
	NoteStore
	TagCategoryStore
	TagStore
	PostStore
	PostMutationLocker
}

// Pinger provides database connectivity checks.
type Pinger interface {
	Ping(ctx context.Context) error
}

// PostMutationLock is a session-level lock held for one post's mutation.
type PostMutationLock interface {
	Unlock(ctx context.Context) error
}

// PostMutationLocker serializes mutations of the same post across store users.
type PostMutationLocker interface {
	AcquirePostMutationLock(ctx context.Context, postID uuid.UUID) (PostMutationLock, error)
}
