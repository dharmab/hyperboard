package async

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/dharmab/hyperboard/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type fileSizePost struct {
	ID         string
	ContentURL string
}

type fileSizeRepository interface {
	NextPostWithoutFileSize(context.Context) (*fileSizePost, error)
	SetPostFileSize(context.Context, string, int64) error
}

type postgresFileSizeRepository struct {
	pool *pgxpool.Pool
}

func (r postgresFileSizeRepository) NextPostWithoutFileSize(ctx context.Context) (*fileSizePost, error) {
	var post fileSizePost
	err := r.pool.QueryRow(ctx,
		"SELECT id::text, content_url FROM posts WHERE file_size IS NULL AND deleted_at IS NULL ORDER BY created_at, id LIMIT 1",
	).Scan(&post.ID, &post.ContentURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &post, nil
}

func (r postgresFileSizeRepository) SetPostFileSize(ctx context.Context, id string, size int64) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE posts SET file_size = $1 WHERE id = $2 AND file_size IS NULL AND deleted_at IS NULL",
		size, id,
	)
	return err
}

type fileSizeController struct {
	repository fileSizeRepository
	mediaStore storage.MediaStore
}

func (c fileSizeController) Reconcile(ctx context.Context) (bool, error) {
	post, err := c.repository.NextPostWithoutFileSize(ctx)
	if err != nil {
		return false, err
	}
	if post == nil {
		return false, nil
	}

	key, err := storageKeyFromURL(post.ContentURL)
	if err != nil {
		return false, fmt.Errorf("determine storage key for post %s: %w", post.ID, err)
	}
	size, err := c.mediaStore.Size(ctx, key)
	if err != nil {
		return false, fmt.Errorf("retrieve file size for post %s: %w", post.ID, err)
	}
	if err := c.repository.SetPostFileSize(ctx, post.ID, size); err != nil {
		return false, fmt.Errorf("store file size for post %s: %w", post.ID, err)
	}
	log.Info().Str("post_id", post.ID).Int64("file_size", size).Msg("backfilled post file size")
	return true, nil
}

func storageKeyFromURL(mediaURL string) (string, error) {
	parsed, err := url.Parse(mediaURL)
	if err != nil {
		return "", err
	}
	const marker = "/posts/"
	index := strings.Index(parsed.Path, marker)
	if index < 0 {
		return "", fmt.Errorf("URL path %q does not contain %q", parsed.Path, marker)
	}
	return strings.TrimPrefix(parsed.Path[index:], "/"), nil
}
