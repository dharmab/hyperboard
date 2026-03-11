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
	"github.com/dharmab/hyperboard/internal/search"
	"github.com/gofrs/uuid/v5"
)

const postColumns = "id, mime_type, content_url, thumbnail_url, note, has_audio, sha256, phash, created_at, updated_at"

func scanPost(row interface{ Scan(...any) error }) (*models.Post, error) {
	var p models.Post
	err := row.Scan(&p.ID, &p.MimeType, &p.ContentURL, &p.ThumbnailURL, &p.Note, &p.HasAudio, &p.Sha256, &p.Phash, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func loadPostTags(ctx context.Context, querier interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}, posts ...*models.Post) error {
	if len(posts) == 0 {
		return nil
	}

	ids := make([]any, len(posts))
	placeholders := make([]string, len(posts))
	postsByID := make(map[uuid.UUID]*models.Post, len(posts))
	for i, p := range posts {
		ids[i] = p.ID
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		postsByID[p.ID] = p
		p.Tags = nil
	}

	query := fmt.Sprintf(
		"SELECT pt.post_id, t.id, t.name FROM posts_tags pt JOIN tags t ON pt.tag_id = t.id WHERE pt.post_id IN (%s) ORDER BY t.name",
		strings.Join(placeholders, ", "),
	)

	rows, err := querier.QueryContext(ctx, query, ids...)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var postID uuid.UUID
		var tag models.Tag
		if err := rows.Scan(&postID, &tag.ID, &tag.Name); err != nil {
			return err
		}
		if p, ok := postsByID[postID]; ok {
			p.Tags = append(p.Tags, &tag)
		}
	}
	return rows.Err()
}

func (s *PostgresSQLStore) ListPosts(ctx context.Context, params ListPostsParams) (models.PostSlice, bool, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = tx.Rollback() }()

	var queryBuilder strings.Builder
	args := []any{}
	paramIdx := 1

	queryBuilder.WriteString("SELECT " + postColumns + " FROM posts WHERE 1=1")

	// Apply tag inclusion filters (direct tags + cascading tags)
	for _, tagName := range params.Query.IncludedTags {
		resolved, err := s.resolveAliasWithTx(ctx, tx, tagName)
		if err != nil {
			return nil, false, err
		}
		fmt.Fprintf(&queryBuilder,
			" AND (EXISTS (SELECT 1 FROM posts_tags JOIN tags ON posts_tags.tag_id = tags.id WHERE posts_tags.post_id = posts.id AND tags.name = $%d)"+
				" OR EXISTS (SELECT 1 FROM posts_tags JOIN tag_cascades ON posts_tags.tag_id = tag_cascades.tag_id JOIN tags ON tag_cascades.cascaded_tag_id = tags.id WHERE posts_tags.post_id = posts.id AND tags.name = $%d))",
			paramIdx, paramIdx,
		)
		args = append(args, resolved)
		paramIdx++
	}

	// Apply tag exclusion filters (direct tags + cascading tags)
	for _, tagName := range params.Query.ExcludedTags {
		resolved, err := s.resolveAliasWithTx(ctx, tx, tagName)
		if err != nil {
			return nil, false, err
		}
		fmt.Fprintf(&queryBuilder,
			" AND NOT EXISTS (SELECT 1 FROM posts_tags JOIN tags ON posts_tags.tag_id = tags.id WHERE posts_tags.post_id = posts.id AND tags.name = $%d)",
			paramIdx,
		)
		args = append(args, resolved)
		paramIdx++
		fmt.Fprintf(&queryBuilder,
			" AND NOT EXISTS (SELECT 1 FROM posts_tags JOIN tag_cascades ON posts_tags.tag_id = tag_cascades.tag_id JOIN tags ON tag_cascades.cascaded_tag_id = tags.id WHERE posts_tags.post_id = posts.id AND tags.name = $%d)",
			paramIdx,
		)
		args = append(args, resolved)
		paramIdx++
	}

	// Apply tagged filter
	if params.Query.Tagged != nil {
		if *params.Query.Tagged {
			queryBuilder.WriteString(" AND EXISTS (SELECT 1 FROM posts_tags WHERE posts_tags.post_id = posts.id)")
		} else {
			queryBuilder.WriteString(" AND NOT EXISTS (SELECT 1 FROM posts_tags WHERE posts_tags.post_id = posts.id)")
		}
	}

	// Apply type filters
	if params.Query.TypeImage {
		queryBuilder.WriteString(" AND mime_type LIKE 'image/%'")
	}
	if params.Query.TypeVideo {
		queryBuilder.WriteString(" AND mime_type LIKE 'video/%'")
	}
	if params.Query.TypeAudio {
		queryBuilder.WriteString(" AND has_audio = true")
	}

	// Random sort
	if params.Query.Sort == search.SortRandom {
		seed := params.RandomSeed
		if seed == nil {
			currentSeed := time.Now().Unix() / 21600
			seed = &currentSeed
		}

		fmt.Fprintf(&queryBuilder, " ORDER BY md5(CAST(posts.id AS text) || $%d)", paramIdx)
		args = append(args, strconv.FormatInt(*seed, 10))
		paramIdx++

		fmt.Fprintf(&queryBuilder, " LIMIT $%d OFFSET $%d", paramIdx, paramIdx+1)
		args = append(args, params.Limit+1, params.RandomOffset)

		rows, err := tx.QueryContext(ctx, queryBuilder.String(), args...)
		if err != nil {
			return nil, false, err
		}
		defer func() { _ = rows.Close() }()

		var posts models.PostSlice
		for rows.Next() {
			p, err := scanPost(rows)
			if err != nil {
				return nil, false, err
			}
			posts = append(posts, p)
		}
		if err := rows.Err(); err != nil {
			return nil, false, err
		}

		if err := loadPostTags(ctx, tx, posts...); err != nil {
			return nil, false, err
		}

		hasMore := len(posts) > params.Limit
		if hasMore {
			posts = posts[:params.Limit]
		}
		return posts, hasMore, nil
	}

	// Deterministic sort (created_at or updated_at)
	sortCol := "created_at"
	if params.Query.Sort == search.SortUpdatedAt {
		sortCol = "updated_at"
	}

	if params.CursorTime != nil && params.CursorID != nil {
		fmt.Fprintf(&queryBuilder,
			" AND (%s < $%d OR (%s = $%d AND id < $%d))",
			sortCol, paramIdx, sortCol, paramIdx, paramIdx+1,
		)
		args = append(args, *params.CursorTime, *params.CursorID)
		paramIdx += 2
	}

	fmt.Fprintf(&queryBuilder, " ORDER BY %s DESC, id DESC", sortCol)
	fmt.Fprintf(&queryBuilder, " LIMIT $%d", paramIdx)
	args = append(args, params.Limit+1)

	rows, err := tx.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = rows.Close() }()

	var posts models.PostSlice
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, false, err
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	if err := loadPostTags(ctx, tx, posts...); err != nil {
		return nil, false, err
	}

	hasMore := len(posts) > params.Limit
	if hasMore {
		posts = posts[:params.Limit]
	}
	return posts, hasMore, nil
}

func (s *PostgresSQLStore) GetPost(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	row := tx.QueryRowContext(ctx, "SELECT "+postColumns+" FROM posts WHERE id = $1", id)
	post, err := scanPost(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if err := loadPostTags(ctx, tx, post); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *PostgresSQLStore) CreatePost(ctx context.Context, input CreatePostInput) (*models.Post, error) {
	row := s.db.QueryRowContext(ctx,
		"INSERT INTO posts (id, mime_type, content_url, thumbnail_url, note, has_audio, sha256, phash, created_at, updated_at) VALUES ($1, $2, $3, $4, '', $5, $6, $7, $8, $9) RETURNING "+postColumns,
		input.ID, input.MimeType, input.ContentURL, input.ThumbnailURL, input.HasAudio, input.Sha256, input.Phash, input.CreatedAt, input.UpdatedAt,
	)
	return scanPost(row)
}

func (s *PostgresSQLStore) UpdatePost(ctx context.Context, id uuid.UUID, note string, tagNames []string, now time.Time) (*models.Post, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// Check that the post exists
	row := tx.QueryRowContext(ctx, "SELECT "+postColumns+" FROM posts WHERE id = $1", id)
	_, err = scanPost(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Update note and updated_at
	row = tx.QueryRowContext(ctx,
		"UPDATE posts SET note = $1, updated_at = $2 WHERE id = $3 RETURNING "+postColumns,
		note, now, id,
	)
	post, err := scanPost(row)
	if err != nil {
		return nil, err
	}

	// Delete all existing tags for this post
	_, err = tx.ExecContext(ctx, "DELETE FROM posts_tags WHERE post_id = $1", id)
	if err != nil {
		return nil, err
	}

	// Insert tags
	for _, tagName := range tagNames {
		resolvedName, resolveErr := s.resolveAliasWithTx(ctx, tx, tagName)
		if resolveErr != nil {
			return nil, resolveErr
		}
		_, err = tx.ExecContext(ctx,
			"INSERT INTO tags (name, created_at, updated_at) VALUES ($1, $2, $3) ON CONFLICT (name) DO NOTHING",
			resolvedName, now, now,
		)
		if err != nil {
			return nil, err
		}

		var tagID uuid.UUID
		err = tx.QueryRowContext(ctx, "SELECT id FROM tags WHERE name = $1", resolvedName).Scan(&tagID)
		if err != nil {
			return nil, err
		}

		_, err = tx.ExecContext(ctx, "INSERT INTO posts_tags (post_id, tag_id) VALUES ($1, $2)", id, tagID)
		if err != nil {
			return nil, err
		}
	}

	if err := loadPostTags(ctx, tx, post); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *PostgresSQLStore) UpdatePostContent(ctx context.Context, id uuid.UUID, input UpdatePostContentInput) (*models.Post, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// Check that the post exists
	row := tx.QueryRowContext(ctx, "SELECT "+postColumns+" FROM posts WHERE id = $1", id)
	_, err = scanPost(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Update content fields
	row = tx.QueryRowContext(ctx,
		"UPDATE posts SET mime_type = $1, content_url = $2, thumbnail_url = $3, has_audio = $4, sha256 = $5, phash = $6, updated_at = $7 WHERE id = $8 RETURNING "+postColumns,
		input.MimeType, input.ContentURL, input.ThumbnailURL, input.HasAudio, input.Sha256, input.Phash, input.UpdatedAt, id,
	)
	post, err := scanPost(row)
	if err != nil {
		return nil, err
	}

	if err := loadPostTags(ctx, tx, post); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *PostgresSQLStore) UpdatePostThumbnail(ctx context.Context, id uuid.UUID, thumbnailURL string, now time.Time) (*models.Post, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// Check that the post exists
	row := tx.QueryRowContext(ctx, "SELECT "+postColumns+" FROM posts WHERE id = $1", id)
	_, err = scanPost(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Update thumbnail
	row = tx.QueryRowContext(ctx,
		"UPDATE posts SET thumbnail_url = $1, updated_at = $2 WHERE id = $3 RETURNING "+postColumns,
		thumbnailURL, now, id,
	)
	post, err := scanPost(row)
	if err != nil {
		return nil, err
	}

	if err := loadPostTags(ctx, tx, post); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *PostgresSQLStore) DeletePost(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	row := tx.QueryRowContext(ctx, "SELECT "+postColumns+" FROM posts WHERE id = $1", id)
	post, err := scanPost(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM posts WHERE id = $1", id)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *PostgresSQLStore) FindPostBySha256(ctx context.Context, hash string) (*models.Post, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+postColumns+" FROM posts WHERE sha256 = $1", hash)
	post, err := scanPost(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return post, nil
}

func (s *PostgresSQLStore) FindSimilarPosts(ctx context.Context, excludeID uuid.UUID, pHash int64, limit int) (models.PostSlice, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var queryBuilder strings.Builder
	args := []any{}
	paramIdx := 1

	queryBuilder.WriteString("SELECT " + postColumns + " FROM posts WHERE phash IS NOT NULL")
	fmt.Fprintf(&queryBuilder, " AND bit_count(CAST((phash # $%d) AS bit(64))) <= $%d", paramIdx, paramIdx+1)
	args = append(args, pHash, s.similarityThreshold)
	paramIdx += 2

	if excludeID != uuid.Nil {
		fmt.Fprintf(&queryBuilder, " AND id != $%d", paramIdx)
		args = append(args, excludeID)
		paramIdx++
	}

	fmt.Fprintf(&queryBuilder, " ORDER BY bit_count(CAST((phash # $%d) AS bit(64)))", 1)
	fmt.Fprintf(&queryBuilder, " LIMIT $%d", paramIdx)
	args = append(args, limit)

	rows, err := tx.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var posts models.PostSlice
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := loadPostTags(ctx, tx, posts...); err != nil {
		return nil, err
	}

	return posts, nil
}

// resolveAliasWithTx resolves an alias using a transaction (for use within transactions).
func (s *PostgresSQLStore) resolveAliasWithTx(ctx context.Context, tx *sql.Tx, name string) (string, error) {
	rows, err := tx.QueryContext(ctx,
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
