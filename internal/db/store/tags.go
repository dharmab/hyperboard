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

// queryable is implemented by both *sql.DB and *sql.Tx.
type queryable interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (s *PostgresSQLStore) ListTags(ctx context.Context, cursor *string, limit int) (models.TagSlice, bool, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = tx.Rollback() }()

	query := `SELECT id, name, description, tag_category_id, created_at, updated_at FROM tags`
	args := []any{}

	if cursor != nil {
		query += ` WHERE name > $1`
		args = append(args, *cursor)
		args = append(args, limit+1)
		query += ` ORDER BY name ASC LIMIT $2`
	} else {
		args = append(args, limit+1)
		query += ` ORDER BY name ASC LIMIT $1`
	}

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = rows.Close() }()

	var tags models.TagSlice
	for rows.Next() {
		t := &models.Tag{}
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.TagCategoryID, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, false, err
		}
		tags = append(tags, t)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	if err := s.loadTagCategories(ctx, tx, tags); err != nil {
		return nil, false, err
	}

	hasMore := len(tags) > limit
	if hasMore {
		tags = tags[:limit]
	}
	return tags, hasMore, nil
}

// loadTagCategories batch-loads tag categories for a slice of tags.
func (s *PostgresSQLStore) loadTagCategories(ctx context.Context, q queryable, tags models.TagSlice) error {
	catIDSet := make(map[uuid.UUID]bool)
	for _, t := range tags {
		if t.TagCategoryID.Valid {
			catIDSet[t.TagCategoryID.V] = true
		}
	}
	if len(catIDSet) == 0 {
		return nil
	}

	catIDs := make([]uuid.UUID, 0, len(catIDSet))
	for id := range catIDSet {
		catIDs = append(catIDs, id)
	}

	args := make([]any, len(catIDs))
	var placeholders strings.Builder
	for i, id := range catIDs {
		if i > 0 {
			placeholders.WriteString(", ")
		}
		placeholders.WriteString("$" + strconv.Itoa(i+1))
		args[i] = id
	}

	rows, err := q.QueryContext(ctx,
		"SELECT id, name, description, color, created_at, updated_at FROM tag_categories WHERE id IN ("+placeholders.String()+")",
		args...,
	)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	catMap := make(map[uuid.UUID]*models.TagCategory)
	for rows.Next() {
		c := &models.TagCategory{}
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.Color, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return err
		}
		catMap[c.ID] = c
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, t := range tags {
		if t.TagCategoryID.Valid {
			t.Category = catMap[t.TagCategoryID.V]
		}
	}
	return nil
}

func (s *PostgresSQLStore) GetTag(ctx context.Context, name string) (*models.Tag, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	t := &models.Tag{}
	err = tx.QueryRowContext(ctx,
		`SELECT id, name, description, tag_category_id, created_at, updated_at FROM tags WHERE name = $1`,
		name,
	).Scan(&t.ID, &t.Name, &t.Description, &t.TagCategoryID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if t.TagCategoryID.Valid {
		if err := s.loadTagCategories(ctx, tx, models.TagSlice{t}); err != nil {
			return nil, err
		}
	}

	return t, nil
}

func (s *PostgresSQLStore) UpsertTag(ctx context.Context, urlName string, input TagInput, now time.Time) (*models.Tag, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = tx.Rollback() }()

	var existing models.Tag
	err = tx.QueryRowContext(ctx,
		`SELECT id, name, description, tag_category_id, created_at, updated_at FROM tags WHERE name = $1`,
		urlName,
	).Scan(&existing.ID, &existing.Name, &existing.Description, &existing.TagCategoryID, &existing.CreatedAt, &existing.UpdatedAt)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, false, err
	}

	var resultModel *models.Tag
	isCreate := errors.Is(err, sql.ErrNoRows)

	if !isCreate {
		updated := &models.Tag{}
		err = tx.QueryRowContext(ctx,
			`UPDATE tags SET name = $1, description = $2, tag_category_id = $3, updated_at = $4 WHERE id = $5
 RETURNING id, name, description, tag_category_id, created_at, updated_at`,
			input.Name, input.Description, input.TagCategoryID, now, existing.ID,
		).Scan(&updated.ID, &updated.Name, &updated.Description, &updated.TagCategoryID, &updated.CreatedAt, &updated.UpdatedAt)
		if err != nil {
			return nil, false, err
		}
		resultModel = updated
	} else {
		id, err := uuid.NewV4()
		if err != nil {
			return nil, false, err
		}
		inserted := &models.Tag{}
		err = tx.QueryRowContext(ctx,
			`INSERT INTO tags (id, name, description, tag_category_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)
 RETURNING id, name, description, tag_category_id, created_at, updated_at`,
			id, input.Name, input.Description, input.TagCategoryID, now, now,
		).Scan(&inserted.ID, &inserted.Name, &inserted.Description, &inserted.TagCategoryID, &inserted.CreatedAt, &inserted.UpdatedAt)
		if err != nil {
			return nil, false, err
		}
		resultModel = inserted
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
		if err := s.loadTagCategories(ctx, s.db, models.TagSlice{resultModel}); err != nil {
			return nil, false, err
		}
	}

	return resultModel, isCreate, nil
}

func (s *PostgresSQLStore) DeleteTag(ctx context.Context, name string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM tags WHERE name = $1`, name)
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

func (s *PostgresSQLStore) GetTagPostCounts(ctx context.Context, tagIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	if len(tagIDs) == 0 {
		return map[uuid.UUID]int{}, nil
	}

	n := len(tagIDs)
	// Use two separate placeholder sets to avoid duplicate $N in the UNION query.
	args := make([]any, n*2)
	var p1, p2 strings.Builder
	for i, id := range tagIDs {
		if i > 0 {
			p1.WriteString(", ")
			p2.WriteString(", ")
		}
		p1.WriteString("$" + strconv.Itoa(i+1))
		p2.WriteString("$" + strconv.Itoa(n+i+1))
		args[i] = id
		args[n+i] = id
	}

	countQuery := `SELECT tag_id, COUNT(DISTINCT post_id) FROM (
SELECT tag_id, post_id FROM posts_tags WHERE tag_id IN (` + p1.String() + `)
UNION ALL
SELECT tc.cascaded_tag_id AS tag_id, pt.post_id
FROM tag_cascades tc
JOIN posts_tags pt ON pt.tag_id = tc.tag_id
WHERE tc.cascaded_tag_id IN (` + p2.String() + `)
) combined GROUP BY tag_id`
	rows, err := s.db.QueryContext(ctx, countQuery, args...)
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
	return s.resolveAliasWithExec(ctx, s.db, name)
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

		resolved, err := s.resolveAliasWithExec(ctx, tx, name)
		if err != nil {
			return err
		}

		var targetID uuid.UUID
		err = tx.QueryRowContext(ctx, "SELECT id FROM tags WHERE name = $1", resolved).Scan(&targetID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("cascade target tag %q not found", name)
			}
			return err
		}

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

	var sourceID uuid.UUID
	err = tx.QueryRowContext(ctx, "SELECT id FROM tags WHERE name = $1", sourceName).Scan(&sourceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("source tag %q: %w", sourceName, ErrNotFound)
		}
		return nil, err
	}

	var targetID uuid.UUID
	err = tx.QueryRowContext(ctx, "SELECT id FROM tags WHERE name = $1", targetName).Scan(&targetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("target tag %q: %w", targetName, ErrNotFound)
		}
		return nil, err
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE posts_tags SET tag_id = $1
 WHERE tag_id = $2
   AND NOT EXISTS (SELECT 1 FROM posts_tags pt2 WHERE pt2.post_id = posts_tags.post_id AND pt2.tag_id = $1)`,
		targetID, sourceID,
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM posts_tags WHERE tag_id = $1", sourceID)
	if err != nil {
		return nil, err
	}

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

	_, err = tx.ExecContext(ctx, "DELETE FROM tags WHERE id = $1", sourceID)
	if err != nil {
		return nil, err
	}

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

	if tagCatID.Valid {
		var cn string
		err = tx.QueryRowContext(ctx, "SELECT name FROM tag_categories WHERE id = $1", tagCatID.V).Scan(&cn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if err == nil {
			tag.Category = &models.TagCategory{Name: cn}
		}
	}

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

// resolveAliasWithExec resolves an alias using a specific executor (for use within transactions).
func (s *PostgresSQLStore) resolveAliasWithExec(ctx context.Context, q queryable, name string) (string, error) {
	rows, err := q.QueryContext(ctx,
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
