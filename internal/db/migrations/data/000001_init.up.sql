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

CREATE TABLE tag_aliases (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tag_id     UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    alias_name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE posts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mime_type     TEXT NOT NULL,
    content_url   TEXT NOT NULL,
    thumbnail_url TEXT NOT NULL,
    note          TEXT NOT NULL DEFAULT '',
    has_audio     BOOLEAN NOT NULL DEFAULT FALSE,
    sha256        TEXT NOT NULL DEFAULT '',
    phash         BIGINT,
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

CREATE INDEX idx_tag_categories_created_at ON tag_categories(created_at);
CREATE INDEX idx_tags_created_at           ON tags(created_at);
CREATE INDEX idx_tag_aliases_tag_id        ON tag_aliases(tag_id);
CREATE INDEX idx_posts_created_at          ON posts(created_at);
CREATE UNIQUE INDEX idx_posts_sha256       ON posts(sha256) WHERE sha256 != '';
CREATE INDEX idx_posts_tags_tag_id         ON posts_tags(tag_id);
CREATE INDEX idx_notes_created_at          ON notes(created_at);

-- Prevent an alias from having the same name as an existing tag
CREATE OR REPLACE FUNCTION check_alias_not_tag_name()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM tags WHERE name = NEW.alias_name) THEN
        RAISE EXCEPTION 'alias "%" conflicts with an existing tag name', NEW.alias_name;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_check_alias_not_tag_name
    BEFORE INSERT OR UPDATE ON tag_aliases
    FOR EACH ROW
    EXECUTE FUNCTION check_alias_not_tag_name();

-- Prevent a tag from being created with the same name as an existing alias
CREATE OR REPLACE FUNCTION check_tag_name_not_alias()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM tag_aliases WHERE alias_name = NEW.name) THEN
        RAISE EXCEPTION 'tag name "%" conflicts with an existing alias', NEW.name;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_check_tag_name_not_alias
    BEFORE INSERT OR UPDATE ON tags
    FOR EACH ROW
    EXECUTE FUNCTION check_tag_name_not_alias();
