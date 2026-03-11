package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/dharmab/hyperboard/internal/db/models"
	"github.com/gofrs/uuid/v5"
)

func (s *PostgresSQLStore) ListNotes(ctx context.Context) (models.NoteSlice, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, content, created_at, updated_at FROM notes ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var notes models.NoteSlice
	for rows.Next() {
		n := &models.Note{}
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func (s *PostgresSQLStore) GetNote(ctx context.Context, id uuid.UUID) (*models.Note, error) {
	n := &models.Note{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, title, content, created_at, updated_at FROM notes WHERE id = $1`,
		id,
	).Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return n, nil
}

func (s *PostgresSQLStore) CreateNote(ctx context.Context, title, content string) (*models.Note, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	n := &models.Note{}
	err = s.db.QueryRowContext(ctx,
		`INSERT INTO notes (id, title, content, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)
 RETURNING id, title, content, created_at, updated_at`,
		id, title, content, now, now,
	).Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (s *PostgresSQLStore) UpdateNote(ctx context.Context, id uuid.UUID, title, content string) (*models.Note, error) {
	now := time.Now().UTC()
	n := &models.Note{}
	err := s.db.QueryRowContext(ctx,
		`UPDATE notes SET title = $1, content = $2, updated_at = $3 WHERE id = $4
 RETURNING id, title, content, created_at, updated_at`,
		title, content, now, id,
	).Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return n, nil
}

func (s *PostgresSQLStore) DeleteNote(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM notes WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
