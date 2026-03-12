package store

import (
	"context"
	"errors"
)

var (
	// ErrNotFound is returned when a requested resource does not exist.
	ErrNotFound = errors.New("not found")
	// ErrAliasConflict is returned when an alias conflicts with an existing tag name.
	ErrAliasConflict = errors.New("alias conflicts with existing tag name")
)

// SQLStore combines all sub-interfaces for database operations.
type SQLStore interface {
	Pinger
	NoteStore
	TagCategoryStore
	TagStore
	PostStore
}

// Pinger provides database connectivity checks.
type Pinger interface {
	Ping(ctx context.Context) error
}
