// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package errors

var TagCascadeErrors = &tagCascadeErrors{
	ErrUniqueTagCascadesPkey: &UniqueConstraintError{
		schema:  "",
		table:   "tag_cascades",
		columns: []string{"tag_id", "cascaded_tag_id"},
		s:       "tag_cascades_pkey",
	},
}

type tagCascadeErrors struct {
	ErrUniqueTagCascadesPkey *UniqueConstraintError
}
