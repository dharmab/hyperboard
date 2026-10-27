DROP INDEX idx_posts_sha256;

CREATE UNIQUE INDEX idx_posts_sha256
    ON posts(sha256)
    WHERE sha256 != '' AND deleted_at IS NULL;
