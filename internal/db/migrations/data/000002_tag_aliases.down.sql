DROP TRIGGER IF EXISTS trg_check_tag_name_not_alias ON tags;
DROP FUNCTION IF EXISTS check_tag_name_not_alias();
DROP TRIGGER IF EXISTS trg_check_alias_not_tag_name ON tag_aliases;
DROP FUNCTION IF EXISTS check_alias_not_tag_name();
DROP TABLE IF EXISTS tag_aliases;
