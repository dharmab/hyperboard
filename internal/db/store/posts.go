package store

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/dharmab/hyperboard/internal/db/models"
	"github.com/dharmab/hyperboard/internal/search"
	"github.com/gofrs/uuid/v5"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

func (s *PostgresSQLStore) ListPosts(ctx context.Context, params ListPostsParams) (models.PostSlice, bool, error) {
	mods := []bob.Mod[*dialect.SelectQuery]{}

	// Apply tag inclusion filters
	for _, tagName := range params.Query.IncludedTags {
		resolved, err := s.ResolveAlias(ctx, tagName)
		if err != nil {
			return nil, false, err
		}
		mods = append(mods, sm.Where(psql.F("EXISTS",
			psql.Select(
				sm.Columns(psql.S("1")),
				sm.From("posts_tags"),
				sm.InnerJoin("tags").OnEQ(models.PostsTags.Columns.TagID, models.Tags.Columns.ID),
				sm.Where(psql.And(
					models.PostsTags.Columns.PostID.EQ(models.Posts.Columns.ID),
					models.Tags.Columns.Name.EQ(psql.Arg(resolved)),
				)),
			),
		)))
	}

	// Apply tag exclusion filters
	for _, tagName := range params.Query.ExcludedTags {
		resolved, err := s.ResolveAlias(ctx, tagName)
		if err != nil {
			return nil, false, err
		}
		mods = append(mods, sm.Where(psql.Not(psql.F("EXISTS",
			psql.Select(
				sm.Columns(psql.S("1")),
				sm.From("posts_tags"),
				sm.InnerJoin("tags").OnEQ(models.PostsTags.Columns.TagID, models.Tags.Columns.ID),
				sm.Where(psql.And(
					models.PostsTags.Columns.PostID.EQ(models.Posts.Columns.ID),
					models.Tags.Columns.Name.EQ(psql.Arg(resolved)),
				)),
			),
		))))
	}

	// Apply tagged: filter
	if params.Query.Tagged != nil {
		if *params.Query.Tagged {
			mods = append(mods, sm.Where(psql.F("EXISTS",
				psql.Select(
					sm.Columns(psql.S("1")),
					sm.From("posts_tags"),
					sm.InnerJoin("tags").OnEQ(models.PostsTags.Columns.TagID, models.Tags.Columns.ID),
					sm.Where(models.PostsTags.Columns.PostID.EQ(models.Posts.Columns.ID)),
				),
			)))
		} else {
			mods = append(mods, sm.Where(psql.Not(psql.F("EXISTS",
				psql.Select(
					sm.Columns(psql.S("1")),
					sm.From("posts_tags"),
					sm.InnerJoin("tags").OnEQ(models.PostsTags.Columns.TagID, models.Tags.Columns.ID),
					sm.Where(models.PostsTags.Columns.PostID.EQ(models.Posts.Columns.ID)),
				),
			))))
		}
	}

	// Apply type filters
	if params.Query.TypeImage {
		mods = append(mods, sm.Where(models.Posts.Columns.MimeType.Like(psql.Arg("image/%"))))
	}
	if params.Query.TypeVideo {
		mods = append(mods, sm.Where(models.Posts.Columns.MimeType.Like(psql.Arg("video/%"))))
	}
	if params.Query.TypeAudio {
		mods = append(mods, sm.Where(models.Posts.Columns.HasAudio.EQ(psql.Arg(true))))
	}

	// Random sort
	if params.Query.Sort == search.SortRandom {
		seed := params.RandomSeed
		if seed == nil {
			currentSeed := time.Now().Unix() / 21600
			seed = &currentSeed
		}

		mods = append(mods,
			sm.OrderBy(dialect.NewFunction("md5",
				psql.Cast(models.Posts.Columns.ID, "text").Concat(psql.Arg(strconv.FormatInt(*seed, 10))),
			)),
			sm.Limit(int64(params.Limit+1)),
			sm.Offset(int64(params.RandomOffset)),
		)

		posts, err := models.Posts.Query(mods...).All(ctx, s.db)
		if err != nil {
			return nil, false, err
		}

		if err := posts.LoadTags(ctx, s.db); err != nil {
			return nil, false, err
		}

		hasMore := len(posts) > params.Limit
		if hasMore {
			posts = posts[:params.Limit]
		}
		return posts, hasMore, nil
	}

	// Deterministic sort (created_at or updated_at)
	sortCol := models.Posts.Columns.CreatedAt
	if params.Query.Sort == search.SortUpdatedAt {
		sortCol = models.Posts.Columns.UpdatedAt
	}
	mods = append(mods,
		sm.OrderBy(sortCol).Desc(),
		sm.OrderBy(models.Posts.Columns.ID).Desc(),
	)

	if params.CursorTime != nil && params.CursorID != nil {
		mods = append(mods, sm.Where(psql.Or(
			sortCol.LT(psql.Arg(*params.CursorTime)),
			psql.And(
				sortCol.EQ(psql.Arg(*params.CursorTime)),
				models.Posts.Columns.ID.LT(psql.Arg(*params.CursorID)),
			),
		)))
	}

	mods = append(mods, sm.Limit(int64(params.Limit+1)))

	posts, err := models.Posts.Query(mods...).All(ctx, s.db)
	if err != nil {
		return nil, false, err
	}

	if err := posts.LoadTags(ctx, s.db); err != nil {
		return nil, false, err
	}

	hasMore := len(posts) > params.Limit
	if hasMore {
		posts = posts[:params.Limit]
	}
	return posts, hasMore, nil
}

func (s *PostgresSQLStore) GetPost(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	model, err := models.Posts.Query(
		sm.Where(models.Posts.Columns.ID.EQ(psql.Arg(id))),
	).One(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if err := model.LoadTags(ctx, s.db); err != nil {
		return nil, err
	}

	return model, nil
}

func (s *PostgresSQLStore) CreatePost(ctx context.Context, setter *models.PostSetter) (*models.Post, error) {
	return models.Posts.Insert(setter).One(ctx, s.db)
}

func (s *PostgresSQLStore) UpdatePost(ctx context.Context, id uuid.UUID, note string, tagNames []string, now time.Time) (*models.Post, error) {
	existingPost, err := models.Posts.Query(
		sm.Where(models.Posts.Columns.ID.EQ(psql.Arg(id))),
	).One(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	nowPtr := &now
	err = existingPost.Update(ctx, tx, &models.PostSetter{
		Note:      &note,
		UpdatedAt: nowPtr,
	})
	if err != nil {
		return nil, err
	}

	_, err = models.PostsTags.Delete(
		dm.Where(models.PostsTags.Columns.PostID.EQ(psql.Arg(id))),
	).Exec(ctx, tx)
	if err != nil {
		return nil, err
	}

	for _, tagName := range tagNames {
		resolvedName, resolveErr := s.resolveAliasWithExec(ctx, tx, tagName)
		if resolveErr != nil {
			return nil, resolveErr
		}
		_, err = tx.ExecContext(ctx,
			"INSERT INTO tags (name, created_at, updated_at) VALUES ($1, $2, $3) ON CONFLICT (name) DO NOTHING",
			resolvedName, nowPtr, nowPtr,
		)
		if err != nil {
			return nil, err
		}
		tag, err := models.Tags.Query(
			sm.Where(models.Tags.Columns.Name.EQ(psql.Arg(resolvedName))),
		).One(ctx, tx)
		if err != nil {
			return nil, err
		}
		err = existingPost.AttachTags(ctx, tx, tag)
		if err != nil {
			return nil, err
		}
	}

	if err := existingPost.LoadTags(ctx, tx); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return existingPost, nil
}

func (s *PostgresSQLStore) UpdatePostContent(ctx context.Context, id uuid.UUID, setter *models.PostSetter) (*models.Post, error) {
	existingPost, err := models.Posts.Query(
		sm.Where(models.Posts.Columns.ID.EQ(psql.Arg(id))),
	).One(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	err = existingPost.Update(ctx, s.db, setter)
	if err != nil {
		return nil, err
	}

	if err = existingPost.LoadTags(ctx, s.db); err != nil {
		return nil, err
	}

	return existingPost, nil
}

func (s *PostgresSQLStore) UpdatePostThumbnail(ctx context.Context, id uuid.UUID, thumbnailURL string, now time.Time) (*models.Post, error) {
	existingPost, err := models.Posts.Query(
		sm.Where(models.Posts.Columns.ID.EQ(psql.Arg(id))),
	).One(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	nowPtr := &now
	err = existingPost.Update(ctx, s.db, &models.PostSetter{
		ThumbnailURL: &thumbnailURL,
		UpdatedAt:    nowPtr,
	})
	if err != nil {
		return nil, err
	}

	if err := existingPost.LoadTags(ctx, s.db); err != nil {
		return nil, err
	}

	return existingPost, nil
}

func (s *PostgresSQLStore) DeletePost(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	post, err := models.Posts.Query(
		sm.Where(models.Posts.Columns.ID.EQ(psql.Arg(id))),
	).One(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	_, err = models.Posts.Delete(
		dm.Where(models.Posts.Columns.ID.EQ(psql.Arg(id))),
	).Exec(ctx, s.db)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (s *PostgresSQLStore) FindPostBySha256(ctx context.Context, hash string) (*models.Post, error) {
	model, err := models.Posts.Query(
		sm.Where(models.Posts.Columns.Sha256.EQ(psql.Arg(hash))),
	).One(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return model, nil
}

func (s *PostgresSQLStore) FindSimilarPosts(ctx context.Context, excludeID uuid.UUID, pHash int64, limit int) (models.PostSlice, error) {
	phashHamming := dialect.NewFunction("bit_count",
		psql.Cast(psql.Group(models.Posts.Columns.Phash.OP("#", psql.Arg(pHash))), "bit(64)"),
	)
	mods := []bob.Mod[*dialect.SelectQuery]{
		sm.Where(models.Posts.Columns.Phash.IsNotNull()),
		sm.Where(phashHamming.LTE(psql.Arg(s.similarityThreshold))),
		sm.OrderBy(phashHamming),
		sm.Limit(int64(limit)),
	}

	if excludeID != uuid.Nil {
		mods = append(mods, sm.Where(models.Posts.Columns.ID.NE(psql.Arg(excludeID))))
	}

	posts, err := models.Posts.Query(mods...).All(ctx, s.db)
	if err != nil {
		return nil, err
	}

	if err := posts.LoadTags(ctx, s.db); err != nil {
		return nil, err
	}

	return posts, nil
}

// resolveAliasWithExec resolves an alias using a specific executor (for use within transactions).
func (s *PostgresSQLStore) resolveAliasWithExec(ctx context.Context, exec bob.Executor, name string) (string, error) {
	rows, err := exec.QueryContext(ctx,
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
