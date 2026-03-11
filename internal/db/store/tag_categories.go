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
)

func (s *PostgresSQLStore) ListTagCategories(ctx context.Context, cursor *string, limit int) (models.TagCategorySlice, bool, error) {
	query := `SELECT id, name, description, color, created_at, updated_at FROM tag_categories`
	args := []any{}

	if cursor != nil {
		query += ` WHERE name > $1`
		args = append(args, *cursor)
	}

	query += ` ORDER BY name ASC LIMIT ` + strconv.Itoa(limit+1)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = rows.Close() }()

	var categories models.TagCategorySlice
	for rows.Next() {
		c := &models.TagCategory{}
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.Color, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, false, err
		}
		categories = append(categories, c)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasMore := len(categories) > limit
	if hasMore {
		categories = categories[:limit]
	}
	return categories, hasMore, nil
}

func (s *PostgresSQLStore) GetTagCategory(ctx context.Context, name string) (*models.TagCategory, error) {
	c := &models.TagCategory{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, description, color, created_at, updated_at FROM tag_categories WHERE name = $1`,
		name,
	).Scan(&c.ID, &c.Name, &c.Description, &c.Color, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return c, nil
}

func (s *PostgresSQLStore) UpsertTagCategory(ctx context.Context, urlName string, input TagCategoryInput, now time.Time) (*models.TagCategory, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = tx.Rollback() }()

	var existing models.TagCategory
	err = tx.QueryRowContext(ctx,
		`SELECT id, name, description, color, created_at, updated_at FROM tag_categories WHERE name = $1`,
		urlName,
	).Scan(&existing.ID, &existing.Name, &existing.Description, &existing.Color, &existing.CreatedAt, &existing.UpdatedAt)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, false, err
	}

	if err == nil {
		updated := &models.TagCategory{}
		err = tx.QueryRowContext(ctx,
			`UPDATE tag_categories SET name = $1, description = $2, color = $3, updated_at = $4 WHERE id = $5
 RETURNING id, name, description, color, created_at, updated_at`,
			input.Name, input.Description, input.Color, now, existing.ID,
		).Scan(&updated.ID, &updated.Name, &updated.Description, &updated.Color, &updated.CreatedAt, &updated.UpdatedAt)
		if err != nil {
			return nil, false, err
		}
		if err := tx.Commit(); err != nil {
			return nil, false, err
		}
		return updated, false, nil
	}

	id, err := uuid.NewV4()
	if err != nil {
		return nil, false, err
	}
	inserted := &models.TagCategory{}
	err = tx.QueryRowContext(ctx,
		`INSERT INTO tag_categories (id, name, description, color, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)
 RETURNING id, name, description, color, created_at, updated_at`,
		id, input.Name, input.Description, input.Color, now, now,
	).Scan(&inserted.ID, &inserted.Name, &inserted.Description, &inserted.Color, &inserted.CreatedAt, &inserted.UpdatedAt)
	if err != nil {
		return nil, false, err
	}
	if err := tx.Commit(); err != nil {
		return nil, false, err
	}
	return inserted, true, nil
}

func (s *PostgresSQLStore) DeleteTagCategory(ctx context.Context, name string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM tag_categories WHERE name = $1`, name)
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
