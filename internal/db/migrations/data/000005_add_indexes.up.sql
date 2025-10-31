CREATE INDEX IF NOT EXISTS idx_tag_categories_created_at ON tag_categories(created_at);
CREATE INDEX IF NOT EXISTS idx_tag_categories_name ON tag_categories(name);
CREATE INDEX IF NOT EXISTS idx_tags_created_at ON tags(created_at);
CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);
CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at);
