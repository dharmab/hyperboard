-- These indexes are redundant because the UNIQUE constraint on the name column
-- already creates an implicit unique index.
DROP INDEX IF EXISTS idx_tags_name;
DROP INDEX IF EXISTS idx_tag_categories_name;
