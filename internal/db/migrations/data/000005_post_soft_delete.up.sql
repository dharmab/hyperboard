ALTER TABLE posts
    ADD COLUMN deleted_at TIMESTAMPTZ;

CREATE INDEX idx_posts_pending_deletion
    ON posts(deleted_at, id)
    WHERE deleted_at IS NOT NULL;
