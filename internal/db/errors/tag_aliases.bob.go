// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package errors

var TagAliasErrors = &tagAliasErrors{
	ErrUniqueTagAliasesPkey: &UniqueConstraintError{
		schema:  "",
		table:   "tag_aliases",
		columns: []string{"id"},
		s:       "tag_aliases_pkey",
	},

	ErrUniqueTagAliasesAliasNameKey: &UniqueConstraintError{
		schema:  "",
		table:   "tag_aliases",
		columns: []string{"alias_name"},
		s:       "tag_aliases_alias_name_key",
	},
}

type tagAliasErrors struct {
	ErrUniqueTagAliasesPkey *UniqueConstraintError

	ErrUniqueTagAliasesAliasNameKey *UniqueConstraintError
}
