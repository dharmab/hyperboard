DROP INDEX IF EXISTS idx_posts_pending_deletion;
ALTER TABLE posts DROP COLUMN IF EXISTS deleted_at;
