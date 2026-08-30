ALTER TABLE posts
    ADD COLUMN file_size BIGINT
    CONSTRAINT posts_file_size_nonnegative CHECK (file_size >= 0);

CREATE INDEX idx_posts_file_size ON posts((COALESCE(file_size, -1)));
