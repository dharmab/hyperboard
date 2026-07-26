package store

import (
	"context"
	"errors"
	"fmt"
)

var (
	// ErrNotFound is returned when a requested resource does not exist.
	ErrNotFound = errors.New("not found")
	// ErrAliasConflict is returned when an alias conflicts with an existing tag name.
	ErrAliasConflict = errors.New("alias conflicts with existing tag name")
)

// CascadeTargetNotFoundError is returned when a cascading tag reference does not resolve to an existing tag.
type CascadeTargetNotFoundError struct {
	// Name is the cascading tag reference as the caller supplied it (before alias resolution).
	Name string
}

func (e *CascadeTargetNotFoundError) Error() string {
	return fmt.Sprintf("cascade target tag %q not found", e.Name)
}

// TagCategoryNameConflictError is returned when creating or renaming a tag category to a name
// already used by another tag category.
type TagCategoryNameConflictError struct {
	// Name is the conflicting tag category name.
	Name string
}

func (e *TagCategoryNameConflictError) Error() string {
	return fmt.Sprintf("tag category name %q already in use", e.Name)
}

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
