package async

import (
	"context"
	"errors"
	"fmt"

	"github.com/dharmab/hyperboard/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type deletedPost struct {
	ID           string
	ContentURL   string
	ThumbnailURL string
}

type deletionRepository interface {
	NextDeletedPost(context.Context) (*deletedPost, error)
	PurgeDeletedPost(context.Context, string) error
}

type postgresDeletionRepository struct {
	pool *pgxpool.Pool
}

func (r postgresDeletionRepository) NextDeletedPost(ctx context.Context) (*deletedPost, error) {
	var post deletedPost
	err := r.pool.QueryRow(ctx,
		"SELECT id::text, content_url, thumbnail_url FROM posts WHERE deleted_at IS NOT NULL ORDER BY deleted_at, id LIMIT 1",
	).Scan(&post.ID, &post.ContentURL, &post.ThumbnailURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &post, nil
}

func (r postgresDeletionRepository) PurgeDeletedPost(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM posts WHERE id = $1 AND deleted_at IS NOT NULL", id)
	return err
}

type deletionController struct {
	repository deletionRepository
	mediaStore storage.MediaStore
}

func (c deletionController) Reconcile(ctx context.Context) (bool, error) {
	post, err := c.repository.NextDeletedPost(ctx)
	if err != nil {
		return false, err
	}
	if post == nil {
		return false, nil
	}

	contentKey, err := storageKeyFromURL(post.ContentURL)
	if err != nil {
		return false, fmt.Errorf("determine content storage key for post %s: %w", post.ID, err)
	}
	thumbnailKey, err := storageKeyFromURL(post.ThumbnailURL)
	if err != nil {
		return false, fmt.Errorf("determine thumbnail storage key for post %s: %w", post.ID, err)
	}
	if err := c.mediaStore.Delete(ctx, contentKey); err != nil {
		return false, fmt.Errorf("delete content for post %s: %w", post.ID, err)
	}
	if err := c.mediaStore.Delete(ctx, thumbnailKey); err != nil {
		return false, fmt.Errorf("delete thumbnail for post %s: %w", post.ID, err)
	}
	if err := c.repository.PurgeDeletedPost(ctx, post.ID); err != nil {
		return false, fmt.Errorf("purge post %s: %w", post.ID, err)
	}

	log.Info().Str("post_id", post.ID).Msg("purged deleted post")
	return true, nil
}
