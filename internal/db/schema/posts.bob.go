// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package schema

import "github.com/aarondl/opt/null"

var Posts = Table[
	postColumns,
	postIndexes,
	postForeignKeys,
	postUniques,
	postChecks,
]{
	Schema: "",
	Name:   "posts",
	Columns: postColumns{
		ID: column{
			Name:      "id",
			DBType:    "uuid",
			Default:   "gen_random_uuid()",
			Comment:   "",
			Nullable:  false,
			Generated: false,
			AutoIncr:  false,
		},
		MimeType: column{
			Name:      "mime_type",
			DBType:    "text",
			Default:   "",
			Comment:   "",
			Nullable:  false,
			Generated: false,
			AutoIncr:  false,
		},
		ContentURL: column{
			Name:      "content_url",
			DBType:    "text",
			Default:   "",
			Comment:   "",
			Nullable:  false,
			Generated: false,
			AutoIncr:  false,
		},
		ThumbnailURL: column{
			Name:      "thumbnail_url",
			DBType:    "text",
			Default:   "",
			Comment:   "",
			Nullable:  false,
			Generated: false,
			AutoIncr:  false,
		},
		Note: column{
			Name:      "note",
			DBType:    "text",
			Default:   "''::text",
			Comment:   "",
			Nullable:  false,
			Generated: false,
			AutoIncr:  false,
		},
		HasAudio: column{
			Name:      "has_audio",
			DBType:    "boolean",
			Default:   "false",
			Comment:   "",
			Nullable:  false,
			Generated: false,
			AutoIncr:  false,
		},
		Sha256: column{
			Name:      "sha256",
			DBType:    "text",
			Default:   "''::text",
			Comment:   "",
			Nullable:  false,
			Generated: false,
			AutoIncr:  false,
		},
		Phash: column{
			Name:      "phash",
			DBType:    "bigint",
			Default:   "NULL",
			Comment:   "",
			Nullable:  true,
			Generated: false,
			AutoIncr:  false,
		},
		CreatedAt: column{
			Name:      "created_at",
			DBType:    "timestamp with time zone",
			Default:   "now()",
			Comment:   "",
			Nullable:  false,
			Generated: false,
			AutoIncr:  false,
		},
		UpdatedAt: column{
			Name:      "updated_at",
			DBType:    "timestamp with time zone",
			Default:   "now()",
			Comment:   "",
			Nullable:  false,
			Generated: false,
			AutoIncr:  false,
		},
	},
	Indexes: postIndexes{
		PostsPkey: index{
			Type: "btree",
			Name: "posts_pkey",
			Columns: []indexColumn{
				{
					Name:         "id",
					Desc:         null.FromCond(false, true),
					IsExpression: false,
				},
			},
			Unique:        true,
			Comment:       "",
			NullsFirst:    []bool{false},
			NullsDistinct: false,
			Where:         "",
			Include:       []string{},
		},
		IdxPostsCreatedAt: index{
			Type: "btree",
			Name: "idx_posts_created_at",
			Columns: []indexColumn{
				{
					Name:         "created_at",
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
		IdxPostsSha256: index{
			Type: "btree",
			Name: "idx_posts_sha256",
			Columns: []indexColumn{
				{
					Name:         "sha256",
					Desc:         null.FromCond(false, true),
					IsExpression: false,
				},
			},
			Unique:        true,
			Comment:       "",
			NullsFirst:    []bool{false},
			NullsDistinct: false,
			Where:         "(sha256 <> ''::text)",
			Include:       []string{},
		},
	},
	PrimaryKey: &constraint{
		Name:    "posts_pkey",
		Columns: []string{"id"},
		Comment: "",
	},

	Comment: "",
}

type postColumns struct {
	ID           column
	MimeType     column
	ContentURL   column
	ThumbnailURL column
	Note         column
	HasAudio     column
	Sha256       column
	Phash        column
	CreatedAt    column
	UpdatedAt    column
}

func (c postColumns) AsSlice() []column {
	return []column{
		c.ID, c.MimeType, c.ContentURL, c.ThumbnailURL, c.Note, c.HasAudio, c.Sha256, c.Phash, c.CreatedAt, c.UpdatedAt,
	}
}

type postIndexes struct {
	PostsPkey         index
	IdxPostsCreatedAt index
	IdxPostsSha256    index
}

func (i postIndexes) AsSlice() []index {
	return []index{
		i.PostsPkey, i.IdxPostsCreatedAt, i.IdxPostsSha256,
	}
}

type postForeignKeys struct{}

func (f postForeignKeys) AsSlice() []foreignKey {
	return []foreignKey{}
}

type postUniques struct{}

func (u postUniques) AsSlice() []constraint {
	return []constraint{}
}

type postChecks struct{}

func (c postChecks) AsSlice() []check {
	return []check{}
}
