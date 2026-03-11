package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/dharmab/hyperboard/internal/db/models"
	"github.com/gofrs/uuid/v5"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

func (s *PostgresSQLStore) ListNotes(ctx context.Context) (models.NoteSlice, error) {
	return models.Notes.Query(
		sm.OrderBy(models.Notes.Columns.CreatedAt).Desc(),
	).All(ctx, s.db)
}

func (s *PostgresSQLStore) GetNote(ctx context.Context, id uuid.UUID) (*models.Note, error) {
	model, err := models.Notes.Query(
		sm.Where(models.Notes.Columns.ID.EQ(psql.Arg(id))),
	).One(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return model, nil
}

func (s *PostgresSQLStore) CreateNote(ctx context.Context, title, content string) (*models.Note, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	now := new(time.Now().UTC())
	return models.Notes.Insert(
		&models.NoteSetter{
			ID:        &id,
			Title:     &title,
			Content:   &content,
			CreatedAt: now,
			UpdatedAt: now,
		},
	).One(ctx, s.db)
}

func (s *PostgresSQLStore) UpdateNote(ctx context.Context, id uuid.UUID, title, content string) (*models.Note, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	model, err := models.Notes.Query(
		sm.Where(models.Notes.Columns.ID.EQ(psql.Arg(id))),
	).One(ctx, tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	now := new(time.Now().UTC())
	err = model.Update(ctx, tx, &models.NoteSetter{
		Title:     &title,
		Content:   &content,
		UpdatedAt: now,
	})
	if err != nil {
		return nil, err
	}
	model.Title = title
	model.Content = content
	model.UpdatedAt = *now

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return model, nil
}

func (s *PostgresSQLStore) DeleteNote(ctx context.Context, id uuid.UUID) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = models.Notes.Query(
		sm.Where(models.Notes.Columns.ID.EQ(psql.Arg(id))),
	).One(ctx, tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	_, err = models.Notes.Delete(
		dm.Where(models.Notes.Columns.ID.EQ(psql.Arg(id))),
	).Exec(ctx, tx)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
