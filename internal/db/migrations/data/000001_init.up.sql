CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE tag_categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    color       TEXT NOT NULL DEFAULT '#888888',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE tags (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name             TEXT NOT NULL UNIQUE,
    description      TEXT NOT NULL DEFAULT '',
    tag_category_id  UUID REFERENCES tag_categories(id) ON DELETE SET NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE posts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mime_type     TEXT NOT NULL,
    content_url   TEXT NOT NULL,
    thumbnail_url TEXT NOT NULL,
    note          TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE posts_tags (
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE ON UPDATE CASCADE,
    tag_id  UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE ON UPDATE CASCADE,
    PRIMARY KEY (post_id, tag_id)
);

CREATE TABLE notes (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title      TEXT NOT NULL,
    content    TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tag_categories_name       ON tag_categories(name);
CREATE INDEX idx_tag_categories_created_at ON tag_categories(created_at);
CREATE INDEX idx_tags_name                 ON tags(name);
CREATE INDEX idx_tags_created_at           ON tags(created_at);
CREATE INDEX idx_posts_created_at          ON posts(created_at);
CREATE INDEX idx_notes_created_at          ON notes(created_at);
