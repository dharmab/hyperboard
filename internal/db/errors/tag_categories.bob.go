// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package errors

var TagCategoryErrors = &tagCategoryErrors{
	ErrUniqueTagCategoriesPkey: &UniqueConstraintError{
		schema:  "",
		table:   "tag_categories",
		columns: []string{"id"},
		s:       "tag_categories_pkey",
	},

	ErrUniqueTagCategoriesNameKey: &UniqueConstraintError{
		schema:  "",
		table:   "tag_categories",
		columns: []string{"name"},
		s:       "tag_categories_name_key",
	},
}

type tagCategoryErrors struct {
	ErrUniqueTagCategoriesPkey *UniqueConstraintError

	ErrUniqueTagCategoriesNameKey *UniqueConstraintError
}
