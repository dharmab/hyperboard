-- Serialize canonical tag-name and alias claims through a shared advisory lock.
-- The lock prevents concurrent transactions from claiming the same name in the
-- tags and tag_aliases tables while retaining the existing table structure.
CREATE OR REPLACE FUNCTION check_alias_not_tag_name()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_advisory_xact_lock(hashtextextended(NEW.alias_name, 0));
    IF EXISTS (SELECT 1 FROM tags WHERE name = NEW.alias_name) THEN
        RAISE unique_violation
            USING MESSAGE = format('alias "%s" conflicts with an existing tag name', NEW.alias_name),
                  CONSTRAINT = 'tag_aliases_alias_name_key';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION check_tag_name_not_alias()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_advisory_xact_lock(hashtextextended(NEW.name, 0));
    IF EXISTS (SELECT 1 FROM tag_aliases WHERE alias_name = NEW.name) THEN
        RAISE unique_violation
            USING MESSAGE = format('tag name "%s" conflicts with an existing alias', NEW.name),
                  CONSTRAINT = 'tags_name_key';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
