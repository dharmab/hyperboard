DROP INDEX IF EXISTS idx_posts_file_size;
ALTER TABLE posts DROP COLUMN IF EXISTS file_size;
