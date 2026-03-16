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

// TagCategoryStore provides CRUD operations for tag categories.
type TagCategoryStore interface {
	// ListTagCategories returns a paginated list of tag categories ordered by name.
	// The bool return indicates whether more results are available beyond the requested limit.
	ListTagCategories(ctx context.Context, cursor *string, limit int) (models.TagCategorySlice, bool, error)
	// GetTagCategory returns a single tag category by name.
	GetTagCategory(ctx context.Context, name string) (*models.TagCategory, error)
	// UpsertTagCategory creates or updates a tag category. urlName is the original name from the URL path (used for renames).
	// The bool return indicates whether a new category was created (true) or an existing one updated (false).
	UpsertTagCategory(ctx context.Context, urlName string, input TagCategoryInput, now time.Time) (*models.TagCategory, bool, error)
	// DeleteTagCategory removes a tag category by name.
	DeleteTagCategory(ctx context.Context, name string) error
	// GetTagCountsByCategory returns the number of tags in each category ID.
	GetTagCountsByCategory(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]int, error)
}

// TagCategoryInput holds the fields for creating or updating a tag category.
type TagCategoryInput struct {
	Name        string
	Description string
	Color       string
}

func (s *PostgresSQLStore) ListTagCategories(ctx context.Context, cursor *string, limit int) (models.TagCategorySlice, bool, error) {
	var args []any
	var query strings.Builder
	query.WriteString("SELECT id, name, description, color, created_at, updated_at FROM tag_categories")

	if cursor != nil {
		query.WriteString(" WHERE name > $1")
		args = append(args, *cursor)
	}

	query.WriteString(" ORDER BY name ASC LIMIT $" + strconv.Itoa(len(args)+1))
	args = append(args, limit+1)

	rows, err := s.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	var categories models.TagCategorySlice
	for rows.Next() {
		cat := &models.TagCategory{}
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.Color, &cat.CreatedAt, &cat.UpdatedAt); err != nil {
			return nil, false, err
		}
		categories = append(categories, cat)
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
	cat := &models.TagCategory{}
	err := s.db.QueryRowContext(ctx,
		"SELECT id, name, description, color, created_at, updated_at FROM tag_categories WHERE name = $1",
		name,
	).Scan(&cat.ID, &cat.Name, &cat.Description, &cat.Color, &cat.CreatedAt, &cat.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return cat, nil
}

func (s *PostgresSQLStore) UpsertTagCategory(ctx context.Context, urlName string, input TagCategoryInput, now time.Time) (*models.TagCategory, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = tx.Rollback() }()

	existing := &models.TagCategory{}
	err = tx.QueryRowContext(ctx,
		"SELECT id, name, description, color, created_at, updated_at FROM tag_categories WHERE name = $1",
		urlName,
	).Scan(&existing.ID, &existing.Name, &existing.Description, &existing.Color, &existing.CreatedAt, &existing.UpdatedAt)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, false, err
	}

	if err == nil {
		_, err = tx.ExecContext(ctx,
			"UPDATE tag_categories SET name = $1, description = $2, color = $3, updated_at = $4 WHERE id = $5",
			input.Name, input.Description, input.Color, now, existing.ID,
		)
		if err != nil {
			return nil, false, err
		}
		existing.Name = input.Name
		existing.Description = input.Description
		existing.Color = input.Color
		existing.UpdatedAt = now

		if err := tx.Commit(); err != nil {
			return nil, false, err
		}
		return existing, false, nil
	}

	inserted := &models.TagCategory{}
	err = tx.QueryRowContext(ctx,
		"INSERT INTO tag_categories (name, description, color, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id, name, description, color, created_at, updated_at",
		input.Name, input.Description, input.Color, now, now,
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
	_, err := s.db.ExecContext(ctx, "DELETE FROM tag_categories WHERE name = $1", name)
	return err
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

	//nolint:gosec // placeholders are parameterized $N values, not user input
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
