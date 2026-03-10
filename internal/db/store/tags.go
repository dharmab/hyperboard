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
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

func (s *PostgresSQLStore) ListTags(ctx context.Context, cursor *string, limit int) (models.TagSlice, bool, error) {
	mods := []bob.Mod[*dialect.SelectQuery]{
		sm.OrderBy(models.Tags.Columns.Name).Asc(),
	}

	if cursor != nil {
		mods = append(mods, sm.Where(models.Tags.Columns.Name.GT(psql.Arg(*cursor))))
	}

	mods = append(mods, sm.Limit(int64(limit+1)))

	tags, err := models.Tags.Query(mods...).All(ctx, s.db)
	if err != nil {
		return nil, false, err
	}

	if err := tags.LoadTagCategory(ctx, s.db); err != nil {
		return nil, false, err
	}

	hasMore := len(tags) > limit
	if hasMore {
		tags = tags[:limit]
	}
	return tags, hasMore, nil
}

func (s *PostgresSQLStore) GetTag(ctx context.Context, name string) (*models.Tag, error) {
	model, err := models.Tags.Query(
		sm.Where(models.Tags.Columns.Name.EQ(psql.Arg(name))),
	).One(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if model.TagCategoryID.Valid {
		if err := model.LoadTagCategory(ctx, s.db); err != nil {
			return nil, err
		}
	}

	return model, nil
}

func (s *PostgresSQLStore) UpsertTag(ctx context.Context, urlName string, input TagInput, now time.Time) (*models.Tag, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	existing, err := models.Tags.Query(
		sm.Where(models.Tags.Columns.Name.EQ(psql.Arg(urlName))),
	).One(ctx, tx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, false, err
	}

	nowPtr := &now
	var resultModel *models.Tag
	isCreate := existing == nil

	if existing != nil {
		err = existing.Update(ctx, tx, &models.TagSetter{
			Name:          &input.Name,
			Description:   &input.Description,
			TagCategoryID: &input.TagCategoryID,
			UpdatedAt:     nowPtr,
		})
		if err != nil {
			return nil, false, err
		}
		existing.Name = input.Name
		existing.Description = input.Description
		existing.TagCategoryID = input.TagCategoryID
		resultModel = existing
	} else {
		resultModel, err = models.Tags.Insert(
			&models.TagSetter{
				Name:          &input.Name,
				Description:   &input.Description,
				TagCategoryID: &input.TagCategoryID,
				CreatedAt:     nowPtr,
				UpdatedAt:     nowPtr,
			},
		).One(ctx, tx)
		if err != nil {
			return nil, false, err
		}
	}

	if err := s.setTagAliases(ctx, tx, resultModel.ID, input.Aliases); err != nil {
		return nil, false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}

	if resultModel.TagCategoryID.Valid {
		if err := resultModel.LoadTagCategory(ctx, s.db); err != nil {
			return nil, false, err
		}
	}

	return resultModel, isCreate, nil
}

func (s *PostgresSQLStore) DeleteTag(ctx context.Context, name string) error {
	_, err := models.Tags.Delete(
		dm.Where(models.Tags.Columns.Name.EQ(psql.Arg(name))),
	).Exec(ctx, s.db)
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

	rows, err := s.db.QueryContext(ctx,
		"SELECT tag_id, COUNT(*) FROM posts_tags WHERE tag_id IN ("+placeholders.String()+") GROUP BY tag_id",
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
func (s *PostgresSQLStore) setTagAliases(ctx context.Context, exec bob.Executor, tagID uuid.UUID, aliases []string) error {
	for _, alias := range aliases {
		if alias == "" {
			continue
		}
		rows, err := exec.QueryContext(ctx, "SELECT COUNT(*) FROM tags WHERE name = $1", alias)
		if err != nil {
			return err
		}
		var count int
		if rows.Next() {
			err = rows.Scan(&count)
		}
		closeErr := rows.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		if count > 0 {
			return fmt.Errorf("%w: %q", ErrAliasConflict, alias)
		}
	}

	_, err := exec.ExecContext(ctx, "DELETE FROM tag_aliases WHERE tag_id = $1", tagID)
	if err != nil {
		return err
	}
	for _, alias := range aliases {
		if alias == "" {
			continue
		}
		_, err := exec.ExecContext(ctx,
			"INSERT INTO tag_aliases (tag_id, alias_name) VALUES ($1, $2)",
			tagID, alias,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresSQLStore) ConvertTagToAlias(ctx context.Context, sourceName, targetName string) (*ConvertTagToAliasResult, error) {
	tx, err := s.db.DB.BeginTx(ctx, nil)
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

	// Load category name if present
	if tagCatID.Valid {
		var cn string
		err = tx.QueryRowContext(ctx, "SELECT name FROM tag_categories WHERE id = $1", tagCatID.V).Scan(&cn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if err == nil {
			tag.R.TagCategory = &models.TagCategory{Name: cn}
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
