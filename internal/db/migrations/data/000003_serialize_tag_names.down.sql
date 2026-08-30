CREATE OR REPLACE FUNCTION check_alias_not_tag_name()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM tags WHERE name = NEW.alias_name) THEN
        RAISE EXCEPTION 'alias "%" conflicts with an existing tag name', NEW.alias_name;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION check_tag_name_not_alias()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM tag_aliases WHERE alias_name = NEW.name) THEN
        RAISE EXCEPTION 'tag name "%" conflicts with an existing alias', NEW.name;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
