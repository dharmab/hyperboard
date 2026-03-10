// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package errors

var NoteErrors = &noteErrors{
	ErrUniqueNotesPkey: &UniqueConstraintError{
		schema:  "",
		table:   "notes",
		columns: []string{"id"},
		s:       "notes_pkey",
	},
}

type noteErrors struct {
	ErrUniqueNotesPkey *UniqueConstraintError
}
