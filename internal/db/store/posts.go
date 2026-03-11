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

func (s *PostgresSQLStore) ListPosts(ctx context.Context, params ListPostsParams) (models.PostSlice, bool, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = tx.Rollback() }()

	var whereParts []string
	var args []any

	argN := func(v any) string {
		args = append(args, v)
		return "$" + strconv.Itoa(len(args))
	}

	for _, tagName := range params.Query.IncludedTags {
		resolved, err := s.resolveAliasWithExec(ctx, tx, tagName)
		if err != nil {
			return nil, false, err
		}
		p1 := argN(resolved)
		p2 := argN(resolved)
		whereParts = append(whereParts, fmt.Sprintf(`(
EXISTS (
SELECT 1 FROM posts_tags pt
JOIN tags t ON pt.tag_id = t.id
WHERE pt.post_id = posts.id AND t.name = %s
)
OR EXISTS (
SELECT 1 FROM posts_tags pt
JOIN tag_cascades tc ON pt.tag_id = tc.tag_id
JOIN tags t ON tc.cascaded_tag_id = t.id
WHERE pt.post_id = posts.id AND t.name = %s
)
)`, p1, p2))
	}

	for _, tagName := range params.Query.ExcludedTags {
		resolved, err := s.resolveAliasWithExec(ctx, tx, tagName)
		if err != nil {
			return nil, false, err
		}
		p1 := argN(resolved)
		p2 := argN(resolved)
		whereParts = append(whereParts, fmt.Sprintf(`NOT EXISTS (
SELECT 1 FROM posts_tags pt
JOIN tags t ON pt.tag_id = t.id
WHERE pt.post_id = posts.id AND t.name = %s
)`, p1))
		whereParts = append(whereParts, fmt.Sprintf(`NOT EXISTS (
SELECT 1 FROM posts_tags pt
JOIN tag_cascades tc ON pt.tag_id = tc.tag_id
JOIN tags t ON tc.cascaded_tag_id = t.id
WHERE pt.post_id = posts.id AND t.name = %s
)`, p2))
	}

	if params.Query.Tagged != nil {
		if *params.Query.Tagged {
			whereParts = append(whereParts, `EXISTS (SELECT 1 FROM posts_tags WHERE post_id = posts.id)`)
		} else {
			whereParts = append(whereParts, `NOT EXISTS (SELECT 1 FROM posts_tags WHERE post_id = posts.id)`)
		}
	}

	if params.Query.TypeImage {
		whereParts = append(whereParts, "mime_type LIKE "+argN("image/%"))
	}
	if params.Query.TypeVideo {
		whereParts = append(whereParts, "mime_type LIKE "+argN("video/%"))
	}
	if params.Query.TypeAudio {
		whereParts = append(whereParts, "has_audio = "+argN(true))
	}

	baseQuery := `SELECT id, mime_type, content_url, thumbnail_url, note, has_audio, sha256, phash, created_at, updated_at FROM posts`

	buildWhereSQL := func() string {
		if len(whereParts) == 0 {
			return ""
		}
		return " WHERE " + strings.Join(whereParts, " AND ")
	}

	if params.Query.Sort == search.SortRandom {
		seed := params.RandomSeed
		if seed == nil {
			currentSeed := time.Now().Unix() / 21600
			seed = &currentSeed
		}

		seedArg := argN(strconv.FormatInt(*seed, 10))
		limitArg := argN(params.Limit + 1)
		offsetArg := argN(params.RandomOffset)

		query := baseQuery + buildWhereSQL() +
			` ORDER BY md5(id::text || ` + seedArg + `) LIMIT ` + limitArg + ` OFFSET ` + offsetArg

		posts, err := s.queryPosts(ctx, tx, query, args...)
		if err != nil {
			return nil, false, err
		}

		if err := s.loadPostTags(ctx, tx, posts); err != nil {
			return nil, false, err
		}

		hasMore := len(posts) > params.Limit
		if hasMore {
			posts = posts[:params.Limit]
		}
		return posts, hasMore, nil
	}

	sortByUpdated := params.Query.Sort == search.SortUpdatedAt

	if params.CursorTime != nil && params.CursorID != nil {
		t1 := argN(*params.CursorTime)
		t2 := argN(*params.CursorTime)
		id1 := argN(*params.CursorID)
		if sortByUpdated {
			whereParts = append(whereParts, `(updated_at < `+t1+` OR (updated_at = `+t2+` AND id < `+id1+`))`)
		} else {
			whereParts = append(whereParts, `(created_at < `+t1+` OR (created_at = `+t2+` AND id < `+id1+`))`)
		}
	}

	limitArg := argN(params.Limit + 1)
	var query string
	if sortByUpdated {
		query = baseQuery + buildWhereSQL() + ` ORDER BY updated_at DESC, id DESC LIMIT ` + limitArg
	} else {
		query = baseQuery + buildWhereSQL() + ` ORDER BY created_at DESC, id DESC LIMIT ` + limitArg
	}

	posts, err := s.queryPosts(ctx, tx, query, args...)
	if err != nil {
		return nil, false, err
	}

	if err := s.loadPostTags(ctx, tx, posts); err != nil {
		return nil, false, err
	}

	hasMore := len(posts) > params.Limit
	if hasMore {
		posts = posts[:params.Limit]
	}
	return posts, hasMore, nil
}

// queryPosts executes a SQL query and scans the results into a PostSlice.
func (s *PostgresSQLStore) queryPosts(ctx context.Context, q queryable, query string, args ...any) (models.PostSlice, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var posts models.PostSlice
	for rows.Next() {
		p := &models.Post{}
		if err := rows.Scan(&p.ID, &p.MimeType, &p.ContentURL, &p.ThumbnailURL, &p.Note, &p.HasAudio, &p.Sha256, &p.Phash, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

// loadPostTags batch-loads tags for a slice of posts.
func (s *PostgresSQLStore) loadPostTags(ctx context.Context, q queryable, posts models.PostSlice) error {
	if len(posts) == 0 {
		return nil
	}

	postIDs := make([]uuid.UUID, len(posts))
	for i, p := range posts {
		postIDs[i] = p.ID
	}

	args := make([]any, len(postIDs))
	var placeholders strings.Builder
	for i, id := range postIDs {
		if i > 0 {
			placeholders.WriteString(", ")
		}
		placeholders.WriteString("$" + strconv.Itoa(i+1))
		args[i] = id
	}

	rows, err := q.QueryContext(ctx,
		`SELECT pt.post_id, t.id, t.name, t.description, t.tag_category_id, t.created_at, t.updated_at
 FROM posts_tags pt
 JOIN tags t ON pt.tag_id = t.id
 WHERE pt.post_id IN (`+placeholders.String()+`)
 ORDER BY t.name`,
		args...,
	)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	tagsByPost := make(map[uuid.UUID]models.TagSlice)
	for rows.Next() {
		var postID uuid.UUID
		t := &models.Tag{}
		if err := rows.Scan(&postID, &t.ID, &t.Name, &t.Description, &t.TagCategoryID, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return err
		}
		tagsByPost[postID] = append(tagsByPost[postID], t)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, p := range posts {
		p.Tags = tagsByPost[p.ID]
	}
	return nil
}

func (s *PostgresSQLStore) GetPost(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	p := &models.Post{}
	err = tx.QueryRowContext(ctx,
		`SELECT id, mime_type, content_url, thumbnail_url, note, has_audio, sha256, phash, created_at, updated_at FROM posts WHERE id = $1`,
		id,
	).Scan(&p.ID, &p.MimeType, &p.ContentURL, &p.ThumbnailURL, &p.Note, &p.HasAudio, &p.Sha256, &p.Phash, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if err := s.loadPostTags(ctx, tx, models.PostSlice{p}); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *PostgresSQLStore) CreatePost(ctx context.Context, input CreatePostInput) (*models.Post, error) {
	p := &models.Post{}
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO posts (id, mime_type, content_url, thumbnail_url, note, has_audio, sha256, phash, created_at, updated_at)
 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
 RETURNING id, mime_type, content_url, thumbnail_url, note, has_audio, sha256, phash, created_at, updated_at`,
		input.ID, input.MimeType, input.ContentURL, input.ThumbnailURL, "", input.HasAudio, input.Sha256, input.Phash, input.CreatedAt, input.UpdatedAt,
	).Scan(&p.ID, &p.MimeType, &p.ContentURL, &p.ThumbnailURL, &p.Note, &p.HasAudio, &p.Sha256, &p.Phash, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PostgresSQLStore) UpdatePost(ctx context.Context, id uuid.UUID, note string, tagNames []string, now time.Time) (*models.Post, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	p := &models.Post{}
	err = tx.QueryRowContext(ctx,
		`UPDATE posts SET note = $1, updated_at = $2 WHERE id = $3
 RETURNING id, mime_type, content_url, thumbnail_url, note, has_audio, sha256, phash, created_at, updated_at`,
		note, now, id,
	).Scan(&p.ID, &p.MimeType, &p.ContentURL, &p.ThumbnailURL, &p.Note, &p.HasAudio, &p.Sha256, &p.Phash, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM posts_tags WHERE post_id = $1", id)
	if err != nil {
		return nil, err
	}

	for _, tagName := range tagNames {
		resolvedName, resolveErr := s.resolveAliasWithExec(ctx, tx, tagName)
		if resolveErr != nil {
			return nil, resolveErr
		}
		_, err = tx.ExecContext(ctx,
			"INSERT INTO tags (id, name, description, created_at, updated_at) VALUES (gen_random_uuid(), $1, '', $2, $3) ON CONFLICT (name) DO NOTHING",
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
		_, err = tx.ExecContext(ctx,
			"INSERT INTO posts_tags (post_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
			id, tagID,
		)
		if err != nil {
			return nil, err
		}
	}

	if err := s.loadPostTags(ctx, tx, models.PostSlice{p}); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *PostgresSQLStore) UpdatePostContent(ctx context.Context, id uuid.UUID, input UpdatePostContentInput) (*models.Post, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	p := &models.Post{}
	err = tx.QueryRowContext(ctx,
		`UPDATE posts SET mime_type = $1, content_url = $2, thumbnail_url = $3, has_audio = $4, sha256 = $5, phash = $6, updated_at = $7 WHERE id = $8
 RETURNING id, mime_type, content_url, thumbnail_url, note, has_audio, sha256, phash, created_at, updated_at`,
		input.MimeType, input.ContentURL, input.ThumbnailURL, input.HasAudio, input.Sha256, input.Phash, input.UpdatedAt, id,
	).Scan(&p.ID, &p.MimeType, &p.ContentURL, &p.ThumbnailURL, &p.Note, &p.HasAudio, &p.Sha256, &p.Phash, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if err := s.loadPostTags(ctx, tx, models.PostSlice{p}); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *PostgresSQLStore) UpdatePostThumbnail(ctx context.Context, id uuid.UUID, thumbnailURL string, now time.Time) (*models.Post, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	p := &models.Post{}
	err = tx.QueryRowContext(ctx,
		`UPDATE posts SET thumbnail_url = $1, updated_at = $2 WHERE id = $3
 RETURNING id, mime_type, content_url, thumbnail_url, note, has_audio, sha256, phash, created_at, updated_at`,
		thumbnailURL, now, id,
	).Scan(&p.ID, &p.MimeType, &p.ContentURL, &p.ThumbnailURL, &p.Note, &p.HasAudio, &p.Sha256, &p.Phash, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if err := s.loadPostTags(ctx, tx, models.PostSlice{p}); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *PostgresSQLStore) DeletePost(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	p := &models.Post{}
	err = tx.QueryRowContext(ctx,
		`SELECT id, mime_type, content_url, thumbnail_url, note, has_audio, sha256, phash, created_at, updated_at FROM posts WHERE id = $1`,
		id,
	).Scan(&p.ID, &p.MimeType, &p.ContentURL, &p.ThumbnailURL, &p.Note, &p.HasAudio, &p.Sha256, &p.Phash, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM posts WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *PostgresSQLStore) FindPostBySha256(ctx context.Context, hash string) (*models.Post, error) {
	p := &models.Post{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, mime_type, content_url, thumbnail_url, note, has_audio, sha256, phash, created_at, updated_at FROM posts WHERE sha256 = $1`,
		hash,
	).Scan(&p.ID, &p.MimeType, &p.ContentURL, &p.ThumbnailURL, &p.Note, &p.HasAudio, &p.Sha256, &p.Phash, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

func (s *PostgresSQLStore) FindSimilarPosts(ctx context.Context, excludeID uuid.UUID, pHash int64, limit int) (models.PostSlice, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var query string
	var queryArgs []any
	if excludeID != uuid.Nil {
		query = `SELECT id, mime_type, content_url, thumbnail_url, note, has_audio, sha256, phash, created_at, updated_at
         FROM posts
         WHERE phash IS NOT NULL
           AND id != $1
           AND bit_count((phash # $2)::bit(64)) <= $3
         ORDER BY bit_count((phash # $2)::bit(64))
         LIMIT $4`
		queryArgs = []any{excludeID, pHash, s.similarityThreshold, limit}
	} else {
		query = `SELECT id, mime_type, content_url, thumbnail_url, note, has_audio, sha256, phash, created_at, updated_at
         FROM posts
         WHERE phash IS NOT NULL
           AND bit_count((phash # $1)::bit(64)) <= $2
         ORDER BY bit_count((phash # $1)::bit(64))
         LIMIT $3`
		queryArgs = []any{pHash, s.similarityThreshold, limit}
	}

	posts, err := s.queryPosts(ctx, tx, query, queryArgs...)
	if err != nil {
		return nil, err
	}

	if err := s.loadPostTags(ctx, tx, posts); err != nil {
		return nil, err
	}

	return posts, nil
}
