package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/dharmab/hyperboard/internal/db/models"
	"github.com/gofrs/uuid/v5"
)

// NoteStore provides CRUD operations for notes.
type NoteStore interface {
	ListNotes(ctx context.Context) (models.NoteSlice, error)
	GetNote(ctx context.Context, id uuid.UUID) (*models.Note, error)
	CreateNote(ctx context.Context, title, content string) (*models.Note, error)
	UpdateNote(ctx context.Context, id uuid.UUID, title, content string) (*models.Note, error)
	DeleteNote(ctx context.Context, id uuid.UUID) error
}

func (s *PostgresSQLStore) ListNotes(ctx context.Context) (models.NoteSlice, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, title, content, created_at, updated_at FROM notes ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes models.NoteSlice
	for rows.Next() {
		note := &models.Note{}
		if err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return notes, nil
}

func (s *PostgresSQLStore) GetNote(ctx context.Context, id uuid.UUID) (*models.Note, error) {
	note := &models.Note{}
	err := s.db.QueryRowContext(ctx, `SELECT id, title, content, created_at, updated_at FROM notes WHERE id = $1`, id).
		Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return note, nil
}

func (s *PostgresSQLStore) CreateNote(ctx context.Context, title, content string) (*models.Note, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()

	note := &models.Note{}
	err = s.db.QueryRowContext(ctx,
		`INSERT INTO notes (id, title, content, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id, title, content, created_at, updated_at`,
		id, title, content, now, now,
	).Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return note, nil
}

func (s *PostgresSQLStore) UpdateNote(ctx context.Context, id uuid.UUID, title, content string) (*models.Note, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var exists uuid.UUID
	err = tx.QueryRowContext(ctx, `SELECT id FROM notes WHERE id = $1`, id).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	now := time.Now().UTC()
	note := &models.Note{}
	err = tx.QueryRowContext(ctx,
		`UPDATE notes SET title = $1, content = $2, updated_at = $3 WHERE id = $4 RETURNING id, title, content, created_at, updated_at`,
		title, content, now, id,
	).Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return note, nil
}

func (s *PostgresSQLStore) DeleteNote(ctx context.Context, id uuid.UUID) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var exists uuid.UUID
	err = tx.QueryRowContext(ctx, `SELECT id FROM notes WHERE id = $1`, id).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM notes WHERE id = $1`, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}
