// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package errors

var TagErrors = &tagErrors{
	ErrUniqueTagsPkey: &UniqueConstraintError{
		schema:  "",
		table:   "tags",
		columns: []string{"id"},
		s:       "tags_pkey",
	},

	ErrUniqueTagsNameKey: &UniqueConstraintError{
		schema:  "",
		table:   "tags",
		columns: []string{"name"},
		s:       "tags_name_key",
	},
}

type tagErrors struct {
	ErrUniqueTagsPkey *UniqueConstraintError

	ErrUniqueTagsNameKey *UniqueConstraintError
}
