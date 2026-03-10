// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package errors

var PostErrors = &postErrors{
	ErrUniquePostsPkey: &UniqueConstraintError{
		schema:  "",
		table:   "posts",
		columns: []string{"id"},
		s:       "posts_pkey",
	},
}

type postErrors struct {
	ErrUniquePostsPkey *UniqueConstraintError
}
