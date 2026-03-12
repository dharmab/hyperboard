package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dharmab/hyperboard/internal/db/models"
	"github.com/gofrs/uuid/v5"
)

// TagStore provides CRUD operations for tags.
type TagStore interface {
	ListTags(ctx context.Context, cursor *string, limit int) (models.TagSlice, bool, error)
	GetTag(ctx context.Context, name string) (*models.Tag, error)
	UpsertTag(ctx context.Context, urlName string, input TagInput, now time.Time) (*models.Tag, bool, error)
	DeleteTag(ctx context.Context, name string) error
	GetTagPostCounts(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]int, error)
	GetTagAliases(ctx context.Context, ids ...uuid.UUID) (map[uuid.UUID][]string, error)
	GetTagCascades(ctx context.Context, ids ...uuid.UUID) (map[uuid.UUID][]string, error)
	ResolveAlias(ctx context.Context, name string) (string, error)
	ConvertTagToAlias(ctx context.Context, sourceName, targetName string) (*ConvertTagToAliasResult, error)
}

// TagInput holds the fields for creating or updating a tag.
type TagInput struct {
	Name          string
	Description   string
	Category      *string
	Aliases       []string
	CascadingTags []string
	TagCategoryID sql.Null[uuid.UUID]
}

// ConvertTagToAliasResult holds the result of converting a tag to an alias.
type ConvertTagToAliasResult struct {
	Tag     *models.Tag
	Aliases []string
}

func (s *PostgresSQLStore) ListTags(ctx context.Context, cursor *string, limit int) (models.TagSlice, bool, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = tx.Rollback() }()

	var args []any
	query := "SELECT id, name, description, tag_category_id, created_at, updated_at FROM tags"
	if cursor != nil {
		query += " WHERE name > $1"
		args = append(args, *cursor)
	}
	//nolint:gosec // placeholders are parameterized $N values, not user input
	query += " ORDER BY name ASC LIMIT $" + strconv.Itoa(len(args)+1)
	args = append(args, limit+1)

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = rows.Close() }()

	var tags models.TagSlice
	categoryIDs := make(map[uuid.UUID]struct{})
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Description, &tag.TagCategoryID, &tag.CreatedAt, &tag.UpdatedAt); err != nil {
			return nil, false, err
		}
		tags = append(tags, &tag)
		if tag.TagCategoryID.Valid {
			categoryIDs[tag.TagCategoryID.V] = struct{}{}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	// Batch-load tag categories
	if len(categoryIDs) > 0 {
		ids := make([]any, 0, len(categoryIDs))
		var placeholders strings.Builder
		i := 0
		for id := range categoryIDs {
			if i > 0 {
				placeholders.WriteString(", ")
			}
			placeholders.WriteString("$" + strconv.Itoa(i+1))
			ids = append(ids, id)
			i++
		}

		//nolint:gosec // placeholders are parameterized $N values, not user input
		catRows, err := tx.QueryContext(ctx,
			"SELECT id, name, description, color, created_at, updated_at FROM tag_categories WHERE id IN ("+placeholders.String()+")",
			ids...,
		)
		if err != nil {
			return nil, false, err
		}
		defer func() { _ = catRows.Close() }()

		categories := make(map[uuid.UUID]*models.TagCategory)
		for catRows.Next() {
			var cat models.TagCategory
			if err := catRows.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.Color, &cat.CreatedAt, &cat.UpdatedAt); err != nil {
				return nil, false, err
			}
			categories[cat.ID] = &cat
		}
		if err := catRows.Err(); err != nil {
			return nil, false, err
		}

		for _, tag := range tags {
			if tag.TagCategoryID.Valid {
				tag.TagCategory = categories[tag.TagCategoryID.V]
			}
		}
	}

	hasMore := len(tags) > limit
	if hasMore {
		tags = tags[:limit]
	}
	return tags, hasMore, nil
}

func (s *PostgresSQLStore) GetTag(ctx context.Context, name string) (*models.Tag, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var tag models.Tag
	err = tx.QueryRowContext(ctx,
		"SELECT id, name, description, tag_category_id, created_at, updated_at FROM tags WHERE name = $1",
		name,
	).Scan(&tag.ID, &tag.Name, &tag.Description, &tag.TagCategoryID, &tag.CreatedAt, &tag.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if tag.TagCategoryID.Valid {
		var cat models.TagCategory
		err = tx.QueryRowContext(ctx,
			"SELECT id, name, description, color, created_at, updated_at FROM tag_categories WHERE id = $1",
			tag.TagCategoryID.V,
		).Scan(&cat.ID, &cat.Name, &cat.Description, &cat.Color, &cat.CreatedAt, &cat.UpdatedAt)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if err == nil {
			tag.TagCategory = &cat
		}
	}

	return &tag, nil
}

func (s *PostgresSQLStore) UpsertTag(ctx context.Context, urlName string, input TagInput, now time.Time) (*models.Tag, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = tx.Rollback() }()

	var existing models.Tag
	err = tx.QueryRowContext(ctx,
		"SELECT id, name, description, tag_category_id, created_at, updated_at FROM tags WHERE name = $1",
		urlName,
	).Scan(&existing.ID, &existing.Name, &existing.Description, &existing.TagCategoryID, &existing.CreatedAt, &existing.UpdatedAt)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, false, err
	}

	var resultModel models.Tag
	isCreate := errors.Is(err, sql.ErrNoRows)

	if !isCreate {
		_, err = tx.ExecContext(ctx,
			"UPDATE tags SET name = $1, description = $2, tag_category_id = $3, updated_at = $4 WHERE id = $5",
			input.Name, input.Description, input.TagCategoryID, now, existing.ID,
		)
		if err != nil {
			return nil, false, err
		}
		resultModel = existing
		resultModel.Name = input.Name
		resultModel.Description = input.Description
		resultModel.TagCategoryID = input.TagCategoryID
		resultModel.UpdatedAt = now
	} else {
		err = tx.QueryRowContext(ctx,
			"INSERT INTO tags (name, description, tag_category_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id, name, description, tag_category_id, created_at, updated_at",
			input.Name, input.Description, input.TagCategoryID, now, now,
		).Scan(&resultModel.ID, &resultModel.Name, &resultModel.Description, &resultModel.TagCategoryID, &resultModel.CreatedAt, &resultModel.UpdatedAt)
		if err != nil {
			return nil, false, err
		}
	}

	if err := s.setTagAliases(ctx, tx, resultModel.ID, input.Aliases); err != nil {
		return nil, false, err
	}

	if err := s.setTagCascades(ctx, tx, resultModel.ID, input.CascadingTags); err != nil {
		return nil, false, err
	}

	if err := tx.Commit(); err != nil {
		return nil, false, err
	}

	if resultModel.TagCategoryID.Valid {
		var cat models.TagCategory
		err = s.db.QueryRowContext(ctx,
			"SELECT id, name, description, color, created_at, updated_at FROM tag_categories WHERE id = $1",
			resultModel.TagCategoryID.V,
		).Scan(&cat.ID, &cat.Name, &cat.Description, &cat.Color, &cat.CreatedAt, &cat.UpdatedAt)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, false, err
		}
		if err == nil {
			resultModel.TagCategory = &cat
		}
	}

	return &resultModel, isCreate, nil
}

func (s *PostgresSQLStore) DeleteTag(ctx context.Context, name string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM tags WHERE name = $1", name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *PostgresSQLStore) GetTagPostCounts(ctx context.Context, tagIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	if len(tagIDs) == 0 {
		return map[uuid.UUID]int{}, nil
	}

	args := make([]any, len(tagIDs))
	var placeholders strings.Builder
	for i, id := range tagIDs {
		if i > 0 {
			placeholders.WriteString(", ")
		}
		placeholders.WriteString("$" + strconv.Itoa(i+1))
		args[i] = id
	}

	// Count direct posts + posts that cascade to each tag
	//nolint:gosec // placeholders are parameterized $N values, not user input
	rows, err := s.db.QueryContext(ctx,
		`SELECT tag_id, COUNT(DISTINCT post_id) FROM (
			SELECT tag_id, post_id FROM posts_tags WHERE tag_id IN (`+placeholders.String()+`)
			UNION ALL
			SELECT tc.cascaded_tag_id AS tag_id, pt.post_id
			FROM tag_cascades tc
			JOIN posts_tags pt ON pt.tag_id = tc.tag_id
			WHERE tc.cascaded_tag_id IN (`+placeholders.String()+`)
		) combined GROUP BY tag_id`,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	counts := make(map[uuid.UUID]int)
	for rows.Next() {
		var tagID uuid.UUID
		var count int
		if err := rows.Scan(&tagID, &count); err != nil {
			return nil, err
		}
		counts[tagID] = count
	}
	return counts, rows.Err()
}

func (s *PostgresSQLStore) GetTagAliases(ctx context.Context, ids ...uuid.UUID) (map[uuid.UUID][]string, error) {
	if len(ids) == 0 {
		return map[uuid.UUID][]string{}, nil
	}

	args := make([]any, len(ids))
	var placeholders strings.Builder
	for i, id := range ids {
		if i > 0 {
			placeholders.WriteString(", ")
		}
		placeholders.WriteString("$" + strconv.Itoa(i+1))
		args[i] = id
	}

	//nolint:gosec // placeholders are parameterized $N values, not user input
	rows, err := s.db.QueryContext(ctx,
		"SELECT tag_id, alias_name FROM tag_aliases WHERE tag_id IN ("+placeholders.String()+") ORDER BY alias_name",
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	result := make(map[uuid.UUID][]string)
	for rows.Next() {
		var tagID uuid.UUID
		var alias string
		if err := rows.Scan(&tagID, &alias); err != nil {
			return nil, err
		}
		result[tagID] = append(result[tagID], alias)
	}
	return result, rows.Err()
}

func (s *PostgresSQLStore) ResolveAlias(ctx context.Context, name string) (string, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT t.name FROM tags t JOIN tag_aliases ta ON t.id = ta.tag_id WHERE ta.alias_name = $1",
		name,
	)
	if err != nil {
		return "", err
	}
	defer func() { _ = rows.Close() }()
	if rows.Next() {
		var canonical string
		if err := rows.Scan(&canonical); err != nil {
			return "", err
		}
		return canonical, rows.Err()
	}
	return name, rows.Err()
}

// setTagAliases replaces all aliases for a tag with the given list.
func (s *PostgresSQLStore) setTagAliases(ctx context.Context, tx *sql.Tx, tagID uuid.UUID, aliases []string) error {
	for _, alias := range aliases {
		if alias == "" {
			continue
		}
		var count int
		err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM tags WHERE name = $1", alias).Scan(&count)
		if err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("%w: %q", ErrAliasConflict, alias)
		}
	}

	_, err := tx.ExecContext(ctx, "DELETE FROM tag_aliases WHERE tag_id = $1", tagID)
	if err != nil {
		return err
	}
	for _, alias := range aliases {
		if alias == "" {
			continue
		}
		_, err := tx.ExecContext(ctx,
			"INSERT INTO tag_aliases (tag_id, alias_name) VALUES ($1, $2)",
			tagID, alias,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresSQLStore) GetTagCascades(ctx context.Context, ids ...uuid.UUID) (map[uuid.UUID][]string, error) {
	if len(ids) == 0 {
		return map[uuid.UUID][]string{}, nil
	}

	args := make([]any, len(ids))
	var placeholders strings.Builder
	for i, id := range ids {
		if i > 0 {
			placeholders.WriteString(", ")
		}
		placeholders.WriteString("$" + strconv.Itoa(i+1))
		args[i] = id
	}

	//nolint:gosec // placeholders are parameterized $N values, not user input
	rows, err := s.db.QueryContext(ctx,
		"SELECT tc.tag_id, t.name FROM tag_cascades tc JOIN tags t ON tc.cascaded_tag_id = t.id WHERE tc.tag_id IN ("+placeholders.String()+") ORDER BY t.name",
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	result := make(map[uuid.UUID][]string)
	for rows.Next() {
		var tagID uuid.UUID
		var name string
		if err := rows.Scan(&tagID, &name); err != nil {
			return nil, err
		}
		result[tagID] = append(result[tagID], name)
	}
	return result, rows.Err()
}

// setTagCascades replaces all cascades for a tag with the given list.
// Each cascade target must be an existing tag (or alias that resolves to one) and must not be the tag itself.
func (s *PostgresSQLStore) setTagCascades(ctx context.Context, tx *sql.Tx, tagID uuid.UUID, cascades []string) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM tag_cascades WHERE tag_id = $1", tagID)
	if err != nil {
		return err
	}
	for _, name := range cascades {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		// Resolve alias to canonical tag name
		resolved, err := s.resolveAliasWithTx(ctx, tx, name)
		if err != nil {
			return err
		}

		// Look up the target tag
		var targetID uuid.UUID
		err = tx.QueryRowContext(ctx, "SELECT id FROM tags WHERE name = $1", resolved).Scan(&targetID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("cascade target tag %q not found", name)
			}
			return err
		}

		// Prevent self-cascade
		if targetID == tagID {
			continue
		}

		_, err = tx.ExecContext(ctx,
			"INSERT INTO tag_cascades (tag_id, cascaded_tag_id) VALUES ($1, $2)",
			tagID, targetID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresSQLStore) GetPostCascadingTags(ctx context.Context, postID uuid.UUID) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT DISTINCT t2.name
		 FROM posts_tags pt
		 JOIN tag_cascades tc ON pt.tag_id = tc.tag_id
		 JOIN tags t2 ON tc.cascaded_tag_id = t2.id
		 WHERE pt.post_id = $1
		   AND t2.id NOT IN (SELECT pt2.tag_id FROM posts_tags pt2 WHERE pt2.post_id = $1)
		 ORDER BY t2.name`,
		postID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		result = append(result, name)
	}
	return result, rows.Err()
}

func (s *PostgresSQLStore) ConvertTagToAlias(ctx context.Context, sourceName, targetName string) (*ConvertTagToAliasResult, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// Look up source tag
	var sourceID uuid.UUID
	err = tx.QueryRowContext(ctx, "SELECT id FROM tags WHERE name = $1", sourceName).Scan(&sourceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("source tag %q: %w", sourceName, ErrNotFound)
		}
		return nil, err
	}

	// Look up target tag
	var targetID uuid.UUID
	err = tx.QueryRowContext(ctx, "SELECT id FROM tags WHERE name = $1", targetName).Scan(&targetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("target tag %q: %w", targetName, ErrNotFound)
		}
		return nil, err
	}

	// Re-tag posts: move source associations to target where target doesn't already exist
	_, err = tx.ExecContext(ctx,
		`UPDATE posts_tags SET tag_id = $1
		 WHERE tag_id = $2
		   AND NOT EXISTS (SELECT 1 FROM posts_tags pt2 WHERE pt2.post_id = posts_tags.post_id AND pt2.tag_id = $1)`,
		targetID, sourceID,
	)
	if err != nil {
		return nil, err
	}

	// Delete remaining source associations
	_, err = tx.ExecContext(ctx, "DELETE FROM posts_tags WHERE tag_id = $1", sourceID)
	if err != nil {
		return nil, err
	}

	// Collect source aliases before deleting
	aliasRows, err := tx.QueryContext(ctx, "SELECT alias_name FROM tag_aliases WHERE tag_id = $1", sourceID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = aliasRows.Close() }()
	var sourceAliases []string
	for aliasRows.Next() {
		var alias string
		if err := aliasRows.Scan(&alias); err != nil {
			return nil, err
		}
		sourceAliases = append(sourceAliases, alias)
	}
	if err := aliasRows.Err(); err != nil {
		return nil, err
	}

	// Delete source tag (cascades aliases)
	_, err = tx.ExecContext(ctx, "DELETE FROM tags WHERE id = $1", sourceID)
	if err != nil {
		return nil, err
	}

	// Add source name + source aliases as aliases of target
	allNewAliases := append([]string{sourceName}, sourceAliases...)
	for _, alias := range allNewAliases {
		_, err = tx.ExecContext(ctx,
			"INSERT INTO tag_aliases (tag_id, alias_name) VALUES ($1, $2) ON CONFLICT DO NOTHING",
			targetID, alias,
		)
		if err != nil {
			return nil, err
		}
	}

	// Read the target tag within the transaction
	var tagName, tagDesc string
	var tagCatID sql.Null[uuid.UUID]
	var tagCreatedAt, tagUpdatedAt time.Time
	err = tx.QueryRowContext(ctx,
		"SELECT name, description, tag_category_id, created_at, updated_at FROM tags WHERE id = $1",
		targetID,
	).Scan(&tagName, &tagDesc, &tagCatID, &tagCreatedAt, &tagUpdatedAt)
	if err != nil {
		return nil, err
	}

	tag := &models.Tag{
		ID:            targetID,
		Name:          tagName,
		Description:   tagDesc,
		TagCategoryID: tagCatID,
		CreatedAt:     tagCreatedAt,
		UpdatedAt:     tagUpdatedAt,
	}

	// Load category if present
	if tagCatID.Valid {
		var cat models.TagCategory
		err = tx.QueryRowContext(ctx,
			"SELECT id, name, description, color, created_at, updated_at FROM tag_categories WHERE id = $1",
			tagCatID.V,
		).Scan(&cat.ID, &cat.Name, &cat.Description, &cat.Color, &cat.CreatedAt, &cat.UpdatedAt)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if err == nil {
			tag.TagCategory = &cat
		}
	}

	// Read aliases
	txAliasRows, err := tx.QueryContext(ctx, "SELECT alias_name FROM tag_aliases WHERE tag_id = $1 ORDER BY alias_name", targetID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = txAliasRows.Close() }()
	var targetAliases []string
	for txAliasRows.Next() {
		var a string
		if err := txAliasRows.Scan(&a); err != nil {
			return nil, err
		}
		targetAliases = append(targetAliases, a)
	}
	if err := txAliasRows.Err(); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &ConvertTagToAliasResult{
		Tag:     tag,
		Aliases: targetAliases,
	}, nil
}
