package store

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/dharmab/hyperboard/internal/db/models"
	"github.com/gofrs/uuid/v5"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

func (s *PostgresSQLStore) ListTagCategories(ctx context.Context, cursor *string, limit int) (models.TagCategorySlice, bool, error) {
	mods := []bob.Mod[*dialect.SelectQuery]{
		sm.OrderBy(models.TagCategories.Columns.Name).Asc(),
	}

	if cursor != nil {
		mods = append(mods, sm.Where(models.TagCategories.Columns.Name.GT(psql.Arg(*cursor))))
	}

	mods = append(mods, sm.Limit(int64(limit+1)))

	categories, err := models.TagCategories.Query(mods...).All(ctx, s.db)
	if err != nil {
		return nil, false, err
	}

	hasMore := len(categories) > limit
	if hasMore {
		categories = categories[:limit]
	}
	return categories, hasMore, nil
}

func (s *PostgresSQLStore) GetTagCategory(ctx context.Context, name string) (*models.TagCategory, error) {
	model, err := models.TagCategories.Query(
		sm.Where(models.TagCategories.Columns.Name.EQ(psql.Arg(name))),
	).One(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return model, nil
}

func (s *PostgresSQLStore) UpsertTagCategory(ctx context.Context, urlName string, input TagCategoryInput, now time.Time) (*models.TagCategory, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	existing, err := models.TagCategories.Query(
		sm.Where(models.TagCategories.Columns.Name.EQ(psql.Arg(urlName))),
	).One(ctx, tx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, false, err
	}

	nowPtr := &now
	if existing != nil {
		err = existing.Update(ctx, tx, &models.TagCategorySetter{
			Name:        &input.Name,
			Description: &input.Description,
			Color:       &input.Color,
			UpdatedAt:   nowPtr,
		})
		if err != nil {
			return nil, false, err
		}
		existing.Name = input.Name
		existing.Description = input.Description
		existing.Color = input.Color
		existing.UpdatedAt = now

		if err := tx.Commit(ctx); err != nil {
			return nil, false, err
		}
		return existing, false, nil
	}

	inserted, err := models.TagCategories.Insert(
		&models.TagCategorySetter{
			Name:        &input.Name,
			Description: &input.Description,
			Color:       &input.Color,
			CreatedAt:   nowPtr,
			UpdatedAt:   nowPtr,
		},
	).One(ctx, tx)
	if err != nil {
		return nil, false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}
	return inserted, true, nil
}

func (s *PostgresSQLStore) DeleteTagCategory(ctx context.Context, name string) error {
	_, err := models.TagCategories.Delete(
		dm.Where(models.TagCategories.Columns.Name.EQ(psql.Arg(name))),
	).Exec(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *PostgresSQLStore) GetTagCountsByCategory(ctx context.Context, categoryIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	if len(categoryIDs) == 0 {
		return map[uuid.UUID]int{}, nil
	}

	args := make([]any, len(categoryIDs))
	var placeholders strings.Builder
	for i, id := range categoryIDs {
		if i > 0 {
			placeholders.WriteString(", ")
		}
		placeholders.WriteString("$" + strconv.Itoa(i+1))
		args[i] = id
	}

	rows, err := s.db.QueryContext(ctx,
		"SELECT tag_category_id, COUNT(*) FROM tags WHERE tag_category_id IN ("+placeholders.String()+") GROUP BY tag_category_id",
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	counts := make(map[uuid.UUID]int)
	for rows.Next() {
		var catID uuid.UUID
		var count int
		if err := rows.Scan(&catID, &count); err != nil {
			return nil, err
		}
		counts[catID] = count
	}
	return counts, rows.Err()
}
