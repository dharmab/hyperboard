CREATE TABLE tag_aliases (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tag_id     UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    alias      TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tag_aliases_tag_id ON tag_aliases(tag_id);
CREATE INDEX idx_tag_aliases_alias ON tag_aliases(alias);

-- Prevent an alias from having the same name as an existing tag
CREATE OR REPLACE FUNCTION check_alias_not_tag_name()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM tags WHERE name = NEW.alias) THEN
        RAISE EXCEPTION 'alias "%" conflicts with an existing tag name', NEW.alias;
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
    IF EXISTS (SELECT 1 FROM tag_aliases WHERE alias = NEW.name) THEN
        RAISE EXCEPTION 'tag name "%" conflicts with an existing alias', NEW.name;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_check_tag_name_not_alias
    BEFORE INSERT OR UPDATE ON tags
    FOR EACH ROW
    EXECUTE FUNCTION check_tag_name_not_alias();
