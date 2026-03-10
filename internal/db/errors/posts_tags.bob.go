// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package errors

var PostsTagErrors = &postsTagErrors{
	ErrUniquePostsTagsPkey: &UniqueConstraintError{
		schema:  "",
		table:   "posts_tags",
		columns: []string{"post_id", "tag_id"},
		s:       "posts_tags_pkey",
	},
}

type postsTagErrors struct {
	ErrUniquePostsTagsPkey *UniqueConstraintError
}
