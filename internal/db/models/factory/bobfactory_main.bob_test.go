// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package factory

import (
	"context"
	"testing"
)

func TestCreateNote(t *testing.T) {
	if testDB == nil {
		t.Skip("skipping test, no DSN provided")
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tx, err := testDB.Begin(ctx)
	if err != nil {
		t.Fatalf("Error starting transaction: %v", err)
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			t.Fatalf("Error rolling back transaction: %v", err)
		}
	}()

	if _, err := New().NewNote(ctx).Create(ctx, tx); err != nil {
		t.Fatalf("Error creating Note: %v", err)
	}
}

func TestCreatePost(t *testing.T) {
	if testDB == nil {
		t.Skip("skipping test, no DSN provided")
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tx, err := testDB.Begin(ctx)
	if err != nil {
		t.Fatalf("Error starting transaction: %v", err)
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			t.Fatalf("Error rolling back transaction: %v", err)
		}
	}()

	if _, err := New().NewPost(ctx).Create(ctx, tx); err != nil {
		t.Fatalf("Error creating Post: %v", err)
	}
}

func TestCreatePostsTag(t *testing.T) {
	if testDB == nil {
		t.Skip("skipping test, no DSN provided")
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tx, err := testDB.Begin(ctx)
	if err != nil {
		t.Fatalf("Error starting transaction: %v", err)
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			t.Fatalf("Error rolling back transaction: %v", err)
		}
	}()

	if _, err := New().NewPostsTag(ctx).Create(ctx, tx); err != nil {
		t.Fatalf("Error creating PostsTag: %v", err)
	}
}

func TestCreateTagAlias(t *testing.T) {
	if testDB == nil {
		t.Skip("skipping test, no DSN provided")
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tx, err := testDB.Begin(ctx)
	if err != nil {
		t.Fatalf("Error starting transaction: %v", err)
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			t.Fatalf("Error rolling back transaction: %v", err)
		}
	}()

	if _, err := New().NewTagAlias(ctx).Create(ctx, tx); err != nil {
		t.Fatalf("Error creating TagAlias: %v", err)
	}
}

func TestCreateTagCategory(t *testing.T) {
	if testDB == nil {
		t.Skip("skipping test, no DSN provided")
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tx, err := testDB.Begin(ctx)
	if err != nil {
		t.Fatalf("Error starting transaction: %v", err)
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			t.Fatalf("Error rolling back transaction: %v", err)
		}
	}()

	if _, err := New().NewTagCategory(ctx).Create(ctx, tx); err != nil {
		t.Fatalf("Error creating TagCategory: %v", err)
	}
}

func TestCreateTag(t *testing.T) {
	if testDB == nil {
		t.Skip("skipping test, no DSN provided")
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tx, err := testDB.Begin(ctx)
	if err != nil {
		t.Fatalf("Error starting transaction: %v", err)
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			t.Fatalf("Error rolling back transaction: %v", err)
		}
	}()

	if _, err := New().NewTag(ctx).Create(ctx, tx); err != nil {
		t.Fatalf("Error creating Tag: %v", err)
	}
}
