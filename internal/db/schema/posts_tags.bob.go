// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package schema

import "github.com/aarondl/opt/null"

var PostsTags = Table[
	postsTagColumns,
	postsTagIndexes,
	postsTagForeignKeys,
	postsTagUniques,
	postsTagChecks,
]{
	Schema: "",
	Name:   "posts_tags",
	Columns: postsTagColumns{
		PostID: column{
			Name:      "post_id",
			DBType:    "uuid",
			Default:   "",
			Comment:   "",
			Nullable:  false,
			Generated: false,
			AutoIncr:  false,
		},
		TagID: column{
			Name:      "tag_id",
			DBType:    "uuid",
			Default:   "",
			Comment:   "",
			Nullable:  false,
			Generated: false,
			AutoIncr:  false,
		},
	},
	Indexes: postsTagIndexes{
		PostsTagsPkey: index{
			Type: "btree",
			Name: "posts_tags_pkey",
			Columns: []indexColumn{
				{
					Name:         "post_id",
					Desc:         null.FromCond(false, true),
					IsExpression: false,
				},
				{
					Name:         "tag_id",
					Desc:         null.FromCond(false, true),
					IsExpression: false,
				},
			},
			Unique:        true,
			Comment:       "",
			NullsFirst:    []bool{false, false},
			NullsDistinct: false,
			Where:         "",
			Include:       []string{},
		},
		IdxPostsTagsTagID: index{
			Type: "btree",
			Name: "idx_posts_tags_tag_id",
			Columns: []indexColumn{
				{
					Name:         "tag_id",
					Desc:         null.FromCond(false, true),
					IsExpression: false,
				},
			},
			Unique:        false,
			Comment:       "",
			NullsFirst:    []bool{false},
			NullsDistinct: false,
			Where:         "",
			Include:       []string{},
		},
	},
	PrimaryKey: &constraint{
		Name:    "posts_tags_pkey",
		Columns: []string{"post_id", "tag_id"},
		Comment: "",
	},
	ForeignKeys: postsTagForeignKeys{
		PostsTagsPostsTagsPostIDFkey: foreignKey{
			constraint: constraint{
				Name:    "posts_tags.posts_tags_post_id_fkey",
				Columns: []string{"post_id"},
				Comment: "",
			},
			ForeignTable:   "posts",
			ForeignColumns: []string{"id"},
		},
		PostsTagsPostsTagsTagIDFkey: foreignKey{
			constraint: constraint{
				Name:    "posts_tags.posts_tags_tag_id_fkey",
				Columns: []string{"tag_id"},
				Comment: "",
			},
			ForeignTable:   "tags",
			ForeignColumns: []string{"id"},
		},
	},

	Comment: "",
}

type postsTagColumns struct {
	PostID column
	TagID  column
}

func (c postsTagColumns) AsSlice() []column {
	return []column{
		c.PostID, c.TagID,
	}
}

type postsTagIndexes struct {
	PostsTagsPkey     index
	IdxPostsTagsTagID index
}

func (i postsTagIndexes) AsSlice() []index {
	return []index{
		i.PostsTagsPkey, i.IdxPostsTagsTagID,
	}
}

type postsTagForeignKeys struct {
	PostsTagsPostsTagsPostIDFkey foreignKey
	PostsTagsPostsTagsTagIDFkey  foreignKey
}

func (f postsTagForeignKeys) AsSlice() []foreignKey {
	return []foreignKey{
		f.PostsTagsPostsTagsPostIDFkey, f.PostsTagsPostsTagsTagIDFkey,
	}
}

type postsTagUniques struct{}

func (u postsTagUniques) AsSlice() []constraint {
	return []constraint{}
}

type postsTagChecks struct{}

func (c postsTagChecks) AsSlice() []check {
	return []check{}
}
